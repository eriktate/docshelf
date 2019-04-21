package bolt

import (
	"context"
	"time"

	"github.com/boltdb/bolt"
	"github.com/docshelf/docshelf"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

// GetGroup fetches an existing docshelf Group from boltdb.
func (s Store) GetGroup(ctx context.Context, id string) (docshelf.Group, error) {
	var group docshelf.Group

	if err := s.fetchItem(ctx, groupBucket, id, &group); err != nil {
		return group, errors.Wrap(err, "failed to fetch group from bolt")
	}

	return group, nil
}

// PutGroup creates a new docshelf Group or updates an existing one in boltdb.
func (s Store) PutGroup(ctx context.Context, group docshelf.Group) (string, error) {
	if group.ID == "" {
		group.ID = xid.New().String()
		group.CreatedAt = time.Now()
	}

	group.UpdatedAt = time.Now()

	if err := s.storeItem(ctx, groupBucket, group.ID, group); err != nil {
		return "", errors.Wrap(err, "failed to put group into bolt")
	}

	return group.ID, nil
}

// RemoveGroup deletes a docshelf Group from boltdb.
func (s Store) RemoveGroup(ctx context.Context, id string) error {
	return errors.Wrap(s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucket)

		if err := b.Delete([]byte(id)); err != nil {
			return err
		}

		return nil
	}), "failed to remove group from bolt")
}
