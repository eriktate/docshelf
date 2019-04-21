package http

import (
	"context"

	"github.com/docshelf/docshelf"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// BasicAuth provides a simple implementation of the Authenticator interface. It does
// a simple lookup of the user and confirms whether or not the credentials match
// what's stored.
type BasicAuth struct {
	userStore docshelf.UserStore
}

// NewBasicAuth returns a new instance of BasicAuth configured with the given
// docshelf.UserStore.
func NewBasicAuth(userStore docshelf.UserStore) BasicAuth {
	return BasicAuth{userStore}
}

// Authenticate implements the docshelf.Authenticator interface. It does a simple pull
// of the user from a UserStore and compares the attempted token with the stored hashed
// token.
func (b BasicAuth) Authenticate(ctx context.Context, email, token string) error {
	user, err := b.userStore.GetUser(ctx, email)
	if err != nil {
		return errors.Wrap(err, "could not find user to authenticate")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Token), []byte(token)); err != nil {
		return errors.New("authentication failed")
	}

	return nil
}
