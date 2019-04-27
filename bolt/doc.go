package bolt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/docshelf/docshelf"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

// GetDoc fetches a docshelf Document from bolt. It will also read and package the Content from an underlying FileStore.
func (s Store) GetDoc(ctx context.Context, path string) (docshelf.Doc, error) {
	var doc docshelf.Doc
	// TODO (erik): This isn't optimized into a single transaction. May want to do that at some point.
	if _, err := xid.FromString(path); err == nil {
		if err := s.fetchItem(ctx, docIDBucket, path, &path); err != nil {
			return doc, err
		}
	}

	if err := s.fetchItem(ctx, docBucket, path, &doc); err != nil {
		if docshelf.CheckDoesNotExist(err) {
			return doc, err
		}

		return doc, errors.Wrap(err, "failed to fetch doc from bolt")
	}

	content, err := s.fs.ReadFile(path)
	if err != nil {
		return doc, errors.Wrap(err, "failed to read file from file store")
	}

	doc.Content = string(content)
	return doc, nil
}

// ListDocs fetches a slice of docshelf Document metadata from bolt. If a query is provided, then the configured
// docshelf.TextIndex will be used to get a set of document paths. If tags are also provided, then they will be used
// to further filter down the results. If no query is provided, but tags are, then the tags will filter down the entire
// set of documents stored.
func (s Store) ListDocs(ctx context.Context, query string, tags ...string) ([]docshelf.Doc, error) {
	var docs []docshelf.Doc
	var foundPaths []string

	// do a full listing if no filters are given
	if query == "" && len(tags) == 0 {
		if err := s.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(docBucket)
			if err := b.ForEach(func(k, v []byte) error {
				var doc docshelf.Doc
				if err := json.Unmarshal(v, &doc); err != nil {
					return err
				}

				docs = append(docs, doc)
				return nil
			}); err != nil {
				return err
			}
			return nil
		}); err != nil {
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

	// short circuit if no tags supplied
	if len(tags) == 0 {
		return s.listDocsTx(ctx, foundPaths)
	}

	var tagged []docshelf.Doc
	if err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		tagged, err = s.listTaggedDocs(ctx, tx, tags)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to list prefix from bolt")
	}

	// short circuit if no query was supplied
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

func (s Store) listDocsTx(ctx context.Context, paths []string) ([]docshelf.Doc, error) {
	var docs []docshelf.Doc
	if err := s.db.View(func(tx *bolt.Tx) error {
		for _, p := range paths {
			var doc docshelf.Doc
			if err := s.getItem(ctx, tx, docBucket, p, &doc); err != nil {
				return err
			}

			docs = append(docs, doc)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return docs, nil
}

func (s Store) listTaggedDocs(ctx context.Context, tx *bolt.Tx, tags []string) ([]docshelf.Doc, error) {
	var docIDs []string
	for _, t := range tags {
		var ids []string
		if err := s.getItem(ctx, tx, tagBucket, t, &ids); err != nil {
			continue
		}

		if docIDs == nil {
			docIDs = ids
		} else {
			docIDs = intersect(docIDs, ids)
		}
	}

	var docs []docshelf.Doc
	for _, id := range docIDs {
		var doc docshelf.Doc
		if err := s.getItem(ctx, tx, docBucket, id, &doc); err != nil {
			continue
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// PutDoc creates or updates an existing docshelf Doc in bolt. It will also store the Content in an underlying FileStore.
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

	// save metadata
	if err := s.db.Update(func(tx *bolt.Tx) error {
		if err := s.putItem(ctx, tx, docBucket, doc.Path, doc); err != nil {

			return errors.Wrap(err, "failed to put doc into bolt")
		}

		if err := s.putItem(ctx, tx, docIDBucket, doc.ID, doc.Path); err != nil {
			return errors.Wrap(err, "failed to save doc secondary index in bolt")
		}

		return nil
	}); err != nil {
		if err := s.fs.RemoveFile(doc.Path); err != nil { // need to rollback file storage if doc fails
			return "", errors.Wrap(err, "failed to put cleanup file after bolt failure")
		}

		return "", err
	}

	return doc.ID, nil
}

// TagDoc tags an existing document with the given tags.
func (s Store) TagDoc(ctx context.Context, path string, tags ...string) error {
	if _, err := xid.FromString(path); err == nil {
		doc, err := s.GetDoc(ctx, path)
		if err != nil {
			return err
		}

		path = doc.Path
	}

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

// RemoveDoc removes a docshelf Doc from bolt as well as the underlying FileStore.
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
