package bolt

import (
	"context"
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
	"github.com/eriktate/skribe"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

var (
	userBucket   = []byte("user")
	groupBucket  = []byte("group")
	docBucket    = []byte("doc")
	policyBucket = []byte("policy")
)

// ErrUserRemoved is a special error case that doesn't necessarily indicate failure.
var ErrUserRemoved = errors.New("user removed in bolt")

// A Store implements several skribe interfaces using boltdb as the backend.
type Store struct {
	db *bolt.DB
}

// New returns a new boltdb Store. This Store can fulfill the interfaces for UserStore, GroupStore, DocStore, and PolicyStore.
func New(db *bolt.DB) (Store, error) {
	if err := db.Update(initBuckets); err != nil {
		return Store{}, err
	}

	return Store{db}, nil
}

// GetUser fetches an existing skribe User from boltdb.
func (s Store) GetUser(ctx context.Context, id string) (skribe.User, error) {
	var user skribe.User
	var val []byte

	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucket)
		val = b.Get([]byte(id))
		return nil
	}); err != nil {
		return user, err
	}

	if err := json.Unmarshal(val, &user); err != nil {
		return user, errors.Wrap(err, "failed to unmarshal user from bolt")
	}

	if user.DeletedAt != nil {
		return skribe.User{}, ErrUserRemoved
	}

	return user, nil
}

// ListUsers returns all skribe Users stored in bolt db.
func (s Store) ListUsers(ctx context.Context) ([]skribe.User, error) {
	users := make([]skribe.User, 0)

	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucket)

		if err := b.ForEach(func(k, v []byte) error {
			var user skribe.User
			if err := json.Unmarshal(v, &user); err != nil {
				return err
			}

			users = append(users, user)

			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return users, nil
}

// PutUser creates a new skribe User or updates an existing one in boltdb.
func (s Store) PutUser(ctx context.Context, user skribe.User) (string, error) {
	if user.ID == "" {
		user.ID = xid.New().String()
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	user.UpdatedAt = time.Now()

	value, err := json.Marshal(&user)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal user for bolt storage")
	}

	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucket)

		if err := b.Put([]byte(user.ID), value); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	return user.ID, nil
}

// RemoveUser marks a user as deleted in boltdb.
func (s Store) RemoveUser(ctx context.Context, id string) error {
	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucket)

		key := []byte(id)
		val := b.Get(key)

		var user skribe.User
		if err := json.Unmarshal(val, &user); err != nil {
			return err
		}

		deletedAt := time.Now()
		user.DeletedAt = &deletedAt
		newVal, err := json.Marshal(&user)
		if err != nil {
			return err
		}

		if err := b.Put(key, newVal); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}
	return nil
}

func initBuckets(tx *bolt.Tx) error {
	if _, err := tx.CreateBucketIfNotExists(userBucket); err != nil {
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

	return nil
}
