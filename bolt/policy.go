package bolt

import (
	"context"
	"time"

	"github.com/boltdb/bolt"
	"github.com/eriktate/docshelf"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

// GetPolicy fetches an existing docshelf Policy from boltdb.
func (s Store) GetPolicy(ctx context.Context, id string) (docshelf.Policy, error) {
	var policy docshelf.Policy

	if err := s.fetchItem(ctx, policyBucket, id, &policy); err != nil {
		return policy, errors.Wrap(err, "failed to fetch policy from bolt")
	}

	return policy, nil
}

// PutPolicy creates a new docshelf Policy or updates an existing one in boltdb.
func (s Store) PutPolicy(ctx context.Context, policy docshelf.Policy) (string, error) {
	if policy.ID == "" {
		policy.ID = xid.New().String()
		policy.CreatedAt = time.Now()
	}

	policy.UpdatedAt = time.Now()

	if err := s.putItem(ctx, policyBucket, policy.ID, policy); err != nil {
		return "", errors.Wrap(err, "failed to put policy into bolt")
	}

	return policy.ID, nil
}

// RemovePolicy deletes a policy from boltdb.
func (s Store) RemovePolicy(ctx context.Context, id string) error {
	return errors.Wrap(s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucket)

		if err := b.Delete([]byte(id)); err != nil {
			return err
		}

		return nil
	}), "failed to remove policy from bolt")
}
