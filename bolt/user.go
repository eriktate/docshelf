package bolt

import (
	"context"
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
	"github.com/eriktate/docshelf"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

// GetUser fetches an existing docshelf User from boltdb.
func (s Store) GetUser(ctx context.Context, id string) (docshelf.User, error) {
	var user docshelf.User

	if err := s.fetchItem(ctx, userBucket, id, &user); err != nil {
		return user, err
	}

	if user.DeletedAt != nil {
		return docshelf.User{}, docshelf.NewErrRemoved("user no longer exists in bolt")
	}

	return user, nil
}

// GetEmail fetches an existing docshelf User from boltdb given an email.
func (s Store) GetEmail(ctx context.Context, id string) (docshelf.User, error) {
	// TODO (erik): Needs to be implemented.
	return docshelf.User{}, errors.New("unimplemented")
}

// ListUsers returns all docshelf Users stored in bolt db.
func (s Store) ListUsers(ctx context.Context) ([]docshelf.User, error) {
	users := make([]docshelf.User, 0)

	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucket)

		if err := b.ForEach(func(k, v []byte) error {
			var user docshelf.User
			if err := json.Unmarshal(v, &user); err != nil {
				return err
			}

			// need to omit deleted users
			if user.DeletedAt != nil {
				return nil
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

// PutUser creates a new docshelf User or updates an existing one in boltdb.
func (s Store) PutUser(ctx context.Context, user docshelf.User) (string, error) {
	if user.ID == "" {
		user.ID = xid.New().String()
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

		var user docshelf.User
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
