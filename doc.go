package docshelf

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docshelf/docshelf"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

// An ID uniquely identifies some entity in docshelf.
type ID string

// NewID generates a new Identifier.
func NewID() ID {
	return ID(xid.New().String())
}

// Identify implements the Identifier interface.
func (i ID) Identify() ID {
	return i
}

// An Identifier is something that can produce an ID.
type Identifier interface {
	Identify() ID
}

// MakeID attempts to create an ID from the given string.
func MakeID(id string) (ID, error) {
	x, err := xid.FromString(id)
	if err != nil {
		return "", errors.New("not a valid ID")
	}

	return ID(x.String()), nil
}

// Meta tracks a lot of information about a document in docshelf.
// This is pretty much everything except for the content itself.
type Meta struct {
	ID        ID
	Title     string
	Path      string
	IsDir     bool
	Parent    ID
	CreatedBy ID
	UpdatedBy ID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Identify implements the Identifier interface.
func (m Meta) Identify() ID {
	return m.ID
}

// A Doc is a full docshelf document.
type Doc struct {
	Meta
	Content string
	Tags    []string
}

// A MetaStore knows how to work with docshelf document meta data.
type MetaStore interface {
	GetMeta(ctx context.Context, id ID) (Meta, error)
	GetMetaByPath(ctx context.Context, path string) (Meta, error)
	ListMeta(ctx context.Context, ids ...ID) ([]Meta, error)
	ListChildren(ctx context.Context, id string) ([]Meta, error)
	SaveMeta(ctx context.Context, meta Meta) (ID, error)
	RemoveMeta(ctx context.Context, id ID) error
}

// A ContentStore knows how to read, write, and remove document content.
type ContentStore interface {
	Read(ctx context.Context, path string) ([]byte, error)
	Write(ctx context.Context, path string, data []byte) error
	Remove(ctx context.Context, path string) error
}

// A Tagger can tag docshelf entities for later retrieval.
type Tagger interface {
	Tag(ctx context.Context, id ID, tags ...string) error
	ListTagged(ctx context.Context, tag string) ([]ID, error)
	ListTags(ctx context.Context, id ID) ([]string, error)
}

// An Indexer can index docshelf documents for searching later.
type Indexer interface {
	Index(ctx context.Context, doc Doc) error
	Search(ctx context.Context, query string) ([]ID, error)
}

// A Service implements all of the business logic required for docshelf to run.
type Service struct {
	metaStore    MetaStore
	contentStore ContentStore
	tagger       Tagger
	indexer      Indexer
}

// GetDoc returns a docshelf Doc given some ID. This can either be a path or a stringified
// identifier.
func (s Service) GetDoc(ctx context.Context, id string) (Doc, error) {
	var meta Meta
	if identifier, err := MakeID(id); err != nil {
		meta, err = s.metaStore.GetMetaByPath(ctx, id)
		if err != nil {
			return Doc{}, err
		}
	} else {
		meta, err = s.metaStore.GetMeta(ctx, identifier)
		if err != nil {
			return Doc{}, err
		}
	}

	content, err := s.contentStore.Read(ctx, meta.Path)
	if err != nil {
		return Doc{}, err
	}

	tags, err := s.tagger.ListTags(ctx, meta.ID)
	if err != nil {
		return Doc{}, err
	}

	return Doc{
		Meta:    meta,
		Content: string(content),
		Tags:    tags,
	}, nil
}

// SaveDoc accepts a docshelf document and either creates it or updates it.
func (s Service) SaveDoc(ctx context.Context, doc Doc) (ID, error) {
	user, err := getContextUser(ctx)
	if err != nil {
		return "", errors.Wrap(err, "could not determine user")
	}

	// CreatedBy ends up getting ignored in the case of updating an existing doc
	doc.CreatedBy = user.ID
	doc.UpdatedBy = user.ID

	// universal requirement
	if doc.Title == "" {
		return "", errors.New("docs must have a title")
	}

	var id ID
	if doc.ID == "" {
		id, err = s.createDoc(ctx, doc)
	} else {
		id, err = s.updateDoc(ctx, doc)
	}

	if err != nil {
		return "", err
	}

	// do indexing here
	if err := s.indexer.Index(ctx, doc); err != nil {
		return id, errors.New("document saved, but indexing failed")
	}

	return id, nil
}

// ListDocs performs a listing of documents
// TODO (erik): This function is a bit wasteful, because it does a full listing at the end
// regardless of whether or not a prefix was supplied.
func (s Service) ListDocs(ctx context.Context, prefix, query string, tags ...string) ([]Meta, error) {
	var ids []ID
	var err error

	// first, get the prefixed meta.
	children, err = s.metaStore.ListChildren(ctx, prefix)
	if err != nil {
		return nil, err
	}

	// second, get the indexed IDs
	if query != "" {
		indexed, err = s.indexer.Search(ctx, query)
		if err != nil {
			return nil, err
		}
		ids = intersect(ids, indexed)
	}

	// third, get the tagged IDs.
	var tagged []ID
	if len(tags) != 0 {
		tagged, err = s.listTagged(tags)
		if err != nil {
			return nil, err
		}
		ids = interesect(ids, tagged)
	}

	// last, find the intersection of those

}

// RemoveDoc removes a document from docshelf.
func (s Service) RemoveDoc(ctx context.Context, id string) error {
	doc, err := s.GetDoc(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to fetch existing document")
	}

	if err := s.contentStore.Remove(ctx, doc.Path); err != nil {
		return errors.Wrap(err, "failed to remove doc content")
	}

	if err := s.metaStore.RemoveMeta(ctx, doc.ID); err != nil {
		return errors.Wrap(err, "failed to remove doc meta")
	}

	// TODO (erik): remove index...?
	// TODO (erik): clean tags...?

	return nil
}

func (s Service) createDoc(ctx context.Context, doc Doc) (ID, error) {
	if doc.ID != "" {
		return "", errors.New("can not create a doc with an existing ID")
	}

	if doc.Path == "" {
		// generate path from title
		doc.Path = strings.Join(strings.Split(strings.ToLower(doc.Title), " "), "-")
	}

	if _, err := s.metaStore.GetMetaByPath(ctx, doc.Path); err != nil {
		if !CheckNotFound(err) {
			return "", err
		}
	} else {
		return "", fmt.Errorf("doc with path %s already exists", doc.Path)
	}

	doc.ID = NewID()
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()

	if err := s.contentStore.Write(ctx, doc.Path, []byte(doc.Content)); err != nil {
		return "", err
	}

	id, err := s.metaStore.SaveMeta(ctx, doc.Meta)
	if err != nil {
		if err = s.contentStore.Remove(ctx, doc.Path); err != nil {
			return "", errors.Wrap(err, "failed to cleanup content on SaveMeta failure")
		}

		return "", err
	}

	if err := s.tagger.Tag(ctx, id, doc.Tags...); err != nil {
		return id, errors.Wrap(err, "doc created, but failed to tag")
	}

	return id, nil
}

func (s Service) updateDoc(ctx context.Context, doc Doc) (ID, error) {
	if doc.ID == "" {
		return "", errors.New("can not update a doc without an ID")
	}

	existing, err := s.metaStore.GetMeta(ctx, doc.ID)
	if err != nil {
		if CheckNotFound(err) {
			return "", fmt.Errorf("doc with ID %s doesn't exist; cannot update", doc.ID)
		}

		return "", errors.Wrap(err, "could not verify existing document")
	}

	if doc.Path == "" {
		doc.Path = existing.Path
	}

	if doc.Path != existing.Path {
		if _, err := s.metaStore.GetMetaByPath(ctx, doc.Path); err != nil {
			if !CheckNotFound(err) {
				return "", errors.Wrap(err, "could not verify existing path")
			}
		} else {
			return "", fmt.Errorf("doc with path %s already exists", doc.Path)
		}
	}

	// automatic fields
	doc.CreatedBy = existing.CreatedBy
	doc.CreatedAt = existing.CreatedAt
	doc.UpdatedAt = time.Now()

	if err := s.contentStore.Write(ctx, doc.Path, []byte(doc.Content)); err != nil {
		return "", err
	}

	if _, err := s.metaStore.SaveMeta(ctx, doc.Meta); err != nil {
		return "", err
	}

	return doc.ID, nil
}

// everything down here is setup for attaching certain data to the request context.
type contextKey string

const userKey = contextKey("ds-user")

func getContextUser(ctx context.Context) (docshelf.User, error) {
	if user, ok := ctx.Value(userKey).(docshelf.User); ok {
		return user, nil
	}

	return docshelf.User{}, errors.New("no user found in context")
}

func intersect(left, right []Identifier) []ID {
	intersection := make([]ID, 0)
	for _, el := range left {
		if contains(right, el) {
			intersection = append(intersection, el.Identify())
		}
	}

	return intersection
}

func contains(slice []Identifier, el Identifier) bool {
	for _, s := range slice {
		if s.Identify() == el.Identify() {
			return true
		}
	}

	return false
}
