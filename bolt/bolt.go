package bolt

import (
	"context"

	"github.com/boltdb/bolt"
	"github.com/eriktate/skribe"
)

// A Store
type Store struct {
	db *bolt.DB
}

func (s Store) GetUser(ctx context.Context, id string) (skribe.User, error) {
}

func (s Store) ListUsers(ctx context.Context) ([]skribe.User, error) {

}

func (s Store) PutUser(ctx context.Context, user skribe.User) (string, error) {

}

func (s Store) RemoveUser(ctx context.Context, id stirng) error {

}
