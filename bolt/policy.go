package bolt

import (
	"context"
	"time"

	"github.com/boltdb/bolt"
	"github.com/eriktate/skribe"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

// GetPolicy fetches an existing skribe Policy from boltdb.
func (s Store) GetPolicy(ctx context.Context, id string) (skribe.Policy, error) {
	var policy skribe.Policy

	if err := s.fetchItem(ctx, policyBucket, id, &policy); err != nil {
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
