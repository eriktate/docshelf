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

	if err := s.getItem(ctx, userBucket, id, &user); err != nil {
		return user, errors.Wrap(err, "failed to fetch user from bolt")
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

	if err := s.putItem(ctx, userBucket, user.ID, user); err != nil {
		return "", errors.Wrap(err, "failed to put user into bolt")
	}

	return user.ID, nil
}

// RemoveUser marks a user as deleted in boltdb.
func (s Store) RemoveUser(ctx context.Context, id string) error {
	return errors.Wrap(s.db.Update(func(tx *bolt.Tx) error {
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
	}), "failed to remove user from bolt")
}

// GetGroup fetches an existing skribe Group from boltdb.
func (s Store) GetGroup(ctx context.Context, id string) (skribe.Group, error) {
	var group skribe.Group

	if err := s.getItem(ctx, groupBucket, id, &group); err != nil {
		return group, errors.Wrap(err, "failed to fetch group from bolt")
	}

	return group, nil
}

// PutGroup creates a new skribe User or updates an existing one in boltdb.
func (s Store) PutGroup(ctx context.Context, group skribe.Group) (string, error) {
	if group.ID == "" {
		group.ID = xid.New().String()
	}

	if group.CreatedAt.IsZero() {
		group.CreatedAt = time.Now()
	}

	group.UpdatedAt = time.Now()

	if err := s.putItem(ctx, groupBucket, group.ID, group); err != nil {
		return "", errors.Wrap(err, "failed to put group into bolt")
	}

	return group.ID, nil
}

// RemoveGroup marks a user as deleted in boltdb.
func (s Store) RemoveGroup(ctx context.Context, id string) error {
	return errors.Wrap(s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucket)

		if err := b.Delete([]byte(id)); err != nil {
			return err
		}

		return nil
	}), "failed to remove group from bolt")
}

// GetPolicy fetches an existing skribe Policy from boltdb.
func (s Store) GetPolicy(ctx context.Context, id string) (skribe.Policy, error) {
	var policy skribe.Policy

	if err := s.getItem(ctx, policyBucket, id, &policy); err != nil {
		return policy, errors.Wrap(err, "failed to fetch policy from bolt")
	}

	return policy, nil
}

// PutPolicy creates a new skribe User or updates an existing one in boltdb.
func (s Store) PutPolicy(ctx context.Context, policy skribe.Policy) (string, error) {
	if policy.ID == "" {
		policy.ID = xid.New().String()
	}

	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = time.Now()
	}

	policy.UpdatedAt = time.Now()

	if err := s.putItem(ctx, policyBucket, policy.ID, policy); err != nil {
		return "", errors.Wrap(err, "failed to put policy into bolt")
	}

	return policy.ID, nil
}

// RemovePolicy marks a user as deleted in boltdb.
func (s Store) RemovePolicy(ctx context.Context, id string) error {
	return errors.Wrap(s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucket)

		if err := b.Delete([]byte(id)); err != nil {
			return err
		}

		return nil
	}), "failed to remove policy from bolt")
}

func (s Store) getItem(ctx context.Context, bucket []byte, id string, out interface{}) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if err := json.Unmarshal(b.Get([]byte(id)), out); err != nil {
			return errors.Wrap(err, "failed to unmarshal entity from bolt")
		}
		return nil
	})
}

func (s Store) putItem(ctx context.Context, bucket []byte, id string, val interface{}) error {
	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		value, err := json.Marshal(val)
		if err != nil {
			return errors.Wrap(err, "failed to marshal entity for bolt storage")
		}

		if err := b.Put([]byte(id), value); err != nil {
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
