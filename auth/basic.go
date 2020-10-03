package auth

import (
	"context"

	"github.com/docshelf/docshelf"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// Basic config for authentication
type Basic struct {
	userStore docshelf.UserStore
}

// NewBasic returns a new Basic authenticator.
func NewBasic(userStore docshelf.UserStore) Basic {
	return Basic{userStore}
}

func (b Basic) Authenticate(ctx context.Context, email, token string) (docshelf.User, error) {
	user, err := b.userStore.GetUser(ctx, email)
	if err != nil {
		return docshelf.User{}, errors.Wrap(err, "could not find user to authenticate")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Token), []byte(token)); err != nil {
		return docshelf.User{}, errors.New("authentication failed")
	}

	return user, nil
}
