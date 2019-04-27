package dynamo

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyna "github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/docshelf/docshelf"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

// A Tag represents the dynamo data structure of a tag.
type Tag struct {
	Tag   string   `json:"tag"`
	Paths []string `json:"paths"`
}

// GetDoc fetches a docshelf Document from dynamodb. It will also read and package the Content
// form an underlying FileStore.
func (s Store) GetDoc(ctx context.Context, path string) (docshelf.Doc, error) {
	var doc docshelf.Doc

	keyName := "path"
	if _, err := xid.FromString(path); err == nil {
		keyName = "id"
	}

	if err := s.getItem(ctx, s.docTable, keyName, path, &doc); err != nil {
		return doc, err
	}

	content, err := s.fs.ReadFile(path)
	if err != nil {
		return doc, err
	}

	doc.Content = string(content)
	return doc, nil
}

// ListDocs fetches a slice of docshelf Document metadata from dynamodb. If a query is provided, then the configured
// docshelf.TextIndex will be used to get a set of document paths. If tags are also provided, then they will be used
// to further filter down the results. If no query is provided, but tags are, then the tags will filter down the entire
// set of documents stored.
func (s Store) ListDocs(ctx context.Context, query string, tags ...string) ([]docshelf.Doc, error) {
	var docs []docshelf.Doc
	var foundPaths []string

	// do a full listing if no filters are given
	if query == "" && len(tags) == 0 {
		input := dynamodb.ScanInput{
			TableName: aws.String(s.docTable),
		}

		res, err := s.client.ScanRequest(&input).Send()
		if err != nil {
			return nil, err
		}

		if err := dyna.UnmarshalListOfMaps(res.Items, &docs); err != nil {
			return nil, err
		}

		return docs, nil
	}

	if query != "" {
		var err error
		foundPaths, err = s.ti.Search(ctx, query)
		if err != nil {
			return nil, err
		}
	}

	if len(tags) == 0 {
		return s.listDocs(ctx, foundPaths)
	}

	tagged, err := s.listTaggedDocs(ctx, tags)
	if err != nil {
		return nil, err
	}

	if query == "" {
		return tagged, nil
	}

	if len(foundPaths) > 0 {
		for _, doc := range tagged {
			if contains(foundPaths, doc.Path) {
				docs = append(docs, doc)
			}
		}
	}

	return docs, nil
}

func (s Store) listDocs(ctx context.Context, paths []string) ([]docshelf.Doc, error) {
	var docs []docshelf.Doc
	for _, path := range paths {
		var doc docshelf.Doc
		if err := s.getItem(ctx, s.docTable, "path", path, &doc); err != nil {
			return nil, err
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

func (s Store) listTaggedDocs(ctx context.Context, tags []string) ([]docshelf.Doc, error) {
	var paths []string
	for _, t := range tags {
		var tag Tag
		if err := s.getItem(ctx, s.tagTable, "tag", t, &tag); err != nil {
			return nil, err
		}

		if paths == nil {
			paths = tag.Paths
		} else {
			paths = intersect(paths, tag.Paths)
		}
	}

	var docs []docshelf.Doc
	for _, path := range paths {
		var doc docshelf.Doc
		if err := s.getItem(ctx, s.docTable, "path", path, &doc); err != nil {
			return nil, err
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// PutDoc creates or updates an existing docshelf Doc in dynamodb. It will also store the
// Content in an underlying FileStore.
func (s Store) PutDoc(ctx context.Context, doc docshelf.Doc) (string, error) {
	if doc.Path == "" {
		return "", errors.New("can not create a new doc without a path")
	}

	if existing, err := s.GetDoc(ctx, doc.Path); err == nil {
		if !docshelf.CheckDoesNotExist(err) {
			return "", errors.Wrap(err, "could not verify existing file")
		}

		doc.CreatedAt = time.Now()
	} else {
		doc.ID = xid.New().String()
		// need to enforce integrity of created* fields if the doc exists.
		doc.CreatedBy = existing.CreatedBy
		doc.CreatedAt = existing.CreatedAt
	}

	doc.UpdatedAt = time.Now()

	// save content
	if err := s.fs.WriteFile(doc.Path, []byte(doc.Content)); err != nil {
		return "", errors.Wrap(err, "failed to write doc to file store")
	}

	// full text index
	if err := s.ti.Index(ctx, doc); err != nil {
		return "", errors.Wrap(err, "failed to text index doc")
	}

	doc.Content = "" // need to clear content before storing doc

	marshaled, err := dyna.MarshalMap(&doc)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal doc for dynamo")
	}

	input := dynamodb.PutItemInput{
		TableName: aws.String(s.docTable),
		Item:      marshaled,
	}

	// save metadata
	if _, err := s.client.PutItemRequest(&input).Send(); err != nil {
		if err := s.fs.RemoveFile(doc.Path); err != nil { // need to rollback file storage if doc failes
			return "", errors.Wrapf(err, "cleanup failed for file: %s", doc.Path)
		}

		return "", errors.Wrap(err, "failed to put doc into dynamo")
	}

	return doc.ID, nil
}

// TagDoc tags an existing document with the given tags.
// TODO (erik): This is a mirror of the bolt implementation. Need to research and find out
// if there's a more efficient way to get this behavior out of dynamo.
func (s Store) TagDoc(ctx context.Context, path string, tags ...string) error {
	if _, err := xid.FromString(path); err == nil {
		doc, err := s.GetDoc(ctx, path)
		if err != nil {
			return err
		}

		path = doc.Path
	}

	for _, t := range tags {
		var tag Tag
		if err := s.getItem(ctx, s.tagTable, "tag", t, &tag); err != nil {
			return err
		}

		// short circuit if the tag alrady contains the path or no tag was returned.
		if contains(tag.Paths, path) {
			continue
		}

		if tag.Tag == "" {
			tag.Tag = t
		}

		tag.Paths = append(tag.Paths, path)
		marshaled, err := dyna.MarshalMap(&tag)
		if err != nil {
			return err
		}

		input := dynamodb.PutItemInput{
			TableName: aws.String(s.tagTable),
			Item:      marshaled,
		}

		if _, err := s.client.PutItemRequest(&input).Send(); err != nil {
			return err
		}
	}

	return nil
}

// RemoveDoc removes a docshelf Doc from dynamo as well as the underlying FileStore.
func (s Store) RemoveDoc(ctx context.Context, path string) error {
	if err := s.fs.RemoveFile(path); err != nil {
		return errors.Wrap(err, "failed to remove doc from file store")
	}

	key, err := makeKey("path", path)
	if err != nil {
		return errors.Wrap(err, "failed to make key")
	}

	input := dynamodb.DeleteItemInput{
		TableName: aws.String(s.docTable),
		Key:       key,
	}

	if _, err := s.client.DeleteItemRequest(&input).Send(); err != nil {
		return errors.Wrap(err, "failed to delete doc from dynamo")
	}

	return nil
}
