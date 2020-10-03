package auth

import (
	"context"
	"fmt"

	"github.com/docshelf/docshelf"
)

// TODO (etate): Identify errors so we only try to PutUser if the user truly wasn't found and it wasn't some random error.
func getOrPutUser(ctx context.Context, store docshelf.UserStore, email, name string) (docshelf.User, error) {
	user, err := store.GetUser(ctx, email)
	if err != nil {
		newUser := docshelf.User{
			Email: email,
			Name:  name,
		}

		id, err := store.PutUser(ctx, newUser)
		if err != nil {
			return user, fmt.Errorf("failed to create new user: %w", err)
		}

		return store.GetUser(ctx, id)
	}

	return user, nil
}
