package bolt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/eriktate/skribe"
	"github.com/pkg/errors"
)

// GetDoc fetches a skribe Document from bolt. It will also read and package the Content from an underlying FileStore.
func (s Store) GetDoc(ctx context.Context, path string) (skribe.Doc, error) {
	var doc skribe.Doc

	if err := s.fetchItem(ctx, docBucket, path, &doc); err != nil {
		if skribe.CheckDoesNotExist(err) {
			return doc, err
		}

		return doc, errors.Wrap(err, "failed to fetch doc from bolt")
	}

	content, err := s.fs.ReadFile(path)
	if err != nil {
		return doc, errors.Wrap(err, "failed to read file from file store")
	}

	doc.Content = content
	return doc, nil
}

// ListDocs fetches a slice of skribe Document metadata from bolt that fit the given prefix or
// tags supplied.
func (s Store) ListDocs(ctx context.Context, prefix string, tags ...string) ([]skribe.Doc, error) {
	var docs []skribe.Doc
	if err := s.db.View(func(tx *bolt.Tx) error {
		// prefer to filter by tag first if supplied.
		if len(tags) > 0 {
			tagged, err := s.listTaggedDocs(ctx, tx, tags)
			if err != nil {
				return err
			}

			// short circuit if no prefix supplied.
			if prefix == "" {
				docs = tagged
				return nil
			}

			listing := make([]skribe.Doc, 0, len(docs))
			for _, doc := range tagged {
				if strings.HasPrefix(doc.Path, prefix) {
					listing = append(listing, doc)
				}
			}

			docs = listing
			return nil
		}

		// if no tags supplied, scan for prefix
		c := tx.Bucket(docBucket).Cursor()

		pre := []byte(prefix)

		// TODO (erik): confirm behavior if prefix is blank.
		for k, v := c.Seek(pre); k != nil && bytes.HasPrefix(k, pre); k, v = c.Next() {
			var doc skribe.Doc
			if err := json.Unmarshal(v, &doc); err != nil {
				return errors.Wrap(err, "failed to unmarshal doc from bolt")
			}

			docs = append(docs, doc)
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to list prefix from bolt")
	}

	return docs, nil
}

func (s Store) listTaggedDocs(ctx context.Context, tx *bolt.Tx, tags []string) ([]skribe.Doc, error) {
	var docIDs []string
	for _, t := range tags {
		var ids []string
		if err := s.getItem(ctx, tx, tagBucket, t, &ids); err != nil {
			return nil, err
		}

		if docIDs == nil {
			docIDs = ids
		} else {
			docIDs = intersect(docIDs, ids)
		}
	}

	var docs []skribe.Doc
	for _, id := range docIDs {
		var doc skribe.Doc
		if err := s.getItem(ctx, tx, tagBucket, id, &doc); err != nil {
			return nil, err
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// PutDoc creates or updates an existing skribe Doc in bolt. It will also store the Content in an underlying FileStore.
func (s Store) PutDoc(ctx context.Context, doc skribe.Doc) error {
	if doc.Path == "" {
		return errors.New("can not create a new doc without a path")
	}

	if _, err := s.GetDoc(ctx, doc.Path); err != nil {
		if !skribe.CheckDoesNotExist(err) {
			return errors.Wrap(err, "could not verify existing file")
		}

		doc.CreatedAt = time.Now()
	}

	doc.UpdatedAt = time.Now()

	if err := s.fs.WriteFile(doc.Path, doc.Content); err != nil {
		return errors.Wrap(err, "failed to write doc to file store")
	}

	doc.Content = nil // need to clear content before storing doc
	if err := s.putItem(ctx, docBucket, doc.Path, doc); err != nil {
		if err := s.fs.RemoveFile(doc.Path); err != nil { // need to rollback file storage if doc fails
			return errors.Wrap(err, "failed to put cleanup file after bolt failure")
		}
		return errors.Wrap(err, "failed to put doc into bolt")
	}

	return nil
}

// TagDoc tags an existing document with the given tags.
func (s Store) TagDoc(ctx context.Context, path string, tags ...string) error {
	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(tagBucket)
		for _, t := range tags {
			val := b.Get([]byte(t))
			if val == nil {
				if err := b.Put([]byte(t), []byte(fmt.Sprintf("[\"%s\"]", path))); err != nil {
					return err
				}

				continue
			}

			var ids []string
			if err := json.Unmarshal(val, &ids); err != nil {
				return err
			}

			if contains(ids, path) {
				continue
			}

			ids = append(ids, path)
			jsonIds, err := json.Marshal(ids)
			if err != nil {
				return err
			}

			if err := b.Put([]byte(t), jsonIds); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to tag document")
	}

	return nil
}

// RemoveDoc removes a skribe Doc from bolt as well as the underlying FileStore.
func (s Store) RemoveDoc(ctx context.Context, path string) error {
	if err := s.fs.RemoveFile(path); err != nil {
		return errors.Wrap(err, "failed to remove doc from file store")
	}

	return errors.Wrap(s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(docBucket)

		if err := b.Delete([]byte(path)); err != nil {
			return err
		}

		return nil
	}), "failed to remove doc from bolt")
}
