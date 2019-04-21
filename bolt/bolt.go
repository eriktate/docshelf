package bolt

import (
	"context"
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/docshelf/docshelf"
	"github.com/pkg/errors"
)

var (
	userBucket      = []byte("user")
	userEmailBucket = []byte("userEmail")
	groupBucket     = []byte("group")
	docBucket       = []byte("doc")
	policyBucket    = []byte("policy")
	tagBucket       = []byte("tag")
)

// A Store implements several docshelf interfaces using boltdb as the backend.
type Store struct {
	db *bolt.DB
	fs docshelf.FileStore
	ti docshelf.TextIndex
}

// New returns a new boltdb Store. This Store can fulfill the interfaces for UserStore, GroupStore, DocStore,
// and PolicyStore.
func New(filename string, fs docshelf.FileStore, ti docshelf.TextIndex) (Store, error) {
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return Store{}, err
	}

	if err := db.Update(initBuckets); err != nil {
		return Store{}, err
	}

	return Store{db, fs, ti}, nil
}

// Close closes the bolt DB file. It currently omits the error for convenience, but that should maybe change
// in the future.
func (s Store) Close() error {
	return s.db.Close()
}

func (s Store) getItem(ctx context.Context, tx *bolt.Tx, bucket []byte, id string, out interface{}) error {
	b := tx.Bucket(bucket)
	val := b.Get([]byte(id))
	if val == nil {
		return docshelf.NewErrDoesNotExist("")
	}

	if err := json.Unmarshal(val, out); err != nil {
		return errors.Wrap(err, "failed to unmarshal entity from bolt")
	}

	return nil
}

func (s Store) fetchItem(ctx context.Context, bucket []byte, id string, out interface{}) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		val := b.Get([]byte(id))
		if val == nil {
			return docshelf.NewErrDoesNotExist("")
		}

		if err := json.Unmarshal(val, out); err != nil {
			return errors.Wrap(err, "failed to unmarshal entity from bolt")
		}

		return nil
	})
}

func (s Store) storeItem(ctx context.Context, bucket []byte, id string, val interface{}) error {
	if err := s.db.Update(func(tx *bolt.Tx) error {
		return s.putItem(ctx, tx, bucket, id, val)
	}); err != nil {
		return err
	}

	return nil
}

func (s Store) putItem(ctx context.Context, tx *bolt.Tx, bucket []byte, id string, val interface{}) error {
	b := tx.Bucket(bucket)
	value, err := json.Marshal(val)
	if err != nil {
		return errors.Wrap(err, "failed to marshal entity for bolt storage")
	}

	if err := b.Put([]byte(id), value); err != nil {
		return err
	}

	return nil
}

func initBuckets(tx *bolt.Tx) error {
	if _, err := tx.CreateBucketIfNotExists(userBucket); err != nil {
		return err
	}

	if _, err := tx.CreateBucketIfNotExists(userEmailBucket); err != nil {
		return err
	}

	if _, err := tx.CreateBucketIfNotExists(groupBucket); err != nil {
		return err
	}

	if _, err := tx.CreateBucketIfNotExists(docBucket); err != nil {
		return err
	}

	if _, err := tx.CreateBucketIfNotExists(policyBucket); err != nil {
		return err
	}

	if _, err := tx.CreateBucketIfNotExists(tagBucket); err != nil {
		return err
	}

	return nil
}

func intersect(left, right []string) []string {
	intersection := make([]string, 0)
	for _, el := range left {
		if contains(right, el) {
			intersection = append(intersection, el)
		}
	}

	return intersection
}

func contains(slice []string, el string) bool {
	for _, s := range slice {
		if s == el {
			return true
		}
	}

	return false
}
