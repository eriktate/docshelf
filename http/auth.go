package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/docshelf/docshelf"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const googleKeyEndpoint = "https://www.googleapis.com/oauth2/v1/certs"

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
	// if no email, assume oauth attempt
	if email == "" {
		if err := b.Validate(ctx, token); err != nil {
			return err
		}
	}

	user, err := b.userStore.GetUser(ctx, email)
	if err != nil {
		return errors.Wrap(err, "could not find user to authenticate")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Token), []byte(token)); err != nil {
		return errors.New("authentication failed")
	}

	return nil
}

func getGooglePublicKeys() (map[string]string, error) {
	client := http.DefaultClient
	res, err := client.Get(googleKeyEndpoint)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New("failed to fetch keys")
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]string)

	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

func (b BasicAuth) Validate(ctx context.Context, token string) error {
	tok, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
		}

		keys, err := getGooglePublicKeys()
		if err != nil {
			return nil, err
		}

		kid := tok.Header["kid"].(string)
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(keys[kid]))
		if err != nil {
			return nil, err
		}

		return publicKey, nil
	})

	log.Printf("Token: %+v", tok)
	if err != nil {
		return fmt.Errorf("failed to parse jwt: %s", err)
	}

	return nil
}
