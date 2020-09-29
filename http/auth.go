package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/docshelf/docshelf"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const googleKeyEndpoint = "https://www.googleapis.com/oauth2/v1/certs"
const googleIssuer = "accounts.google.com"

type Claims struct {
	jwt.StandardClaims

	Email string `json:"email"`
}

// Auth provides a simple implementation of the Authenticator interface. It does
// a simple lookup of the user and confirms whether or not the credentials match
// what's stored.
type Auth struct {
	userStore docshelf.UserStore
}

// NewAuth returns a new instance of Auth configured with the given
// docshelf.UserStore.
func NewAuth(userStore docshelf.UserStore) Auth {
	return Auth{userStore}
}

// Authenticate implements the docshelf.Authenticator interface. It does a simple pull
// of the user from a UserStore and compares the attempted token with the stored hashed
// token.
func (a Auth) Authenticate(ctx context.Context, email, token string) error {
	// if no email, assume oauth attempt
	if email == "" {
		if err := a.Validate(ctx, token); err != nil {
			return err
		}
	}

	// if email and token combination are provided, attempt basic auth
	user, err := a.userStore.GetUser(ctx, email)
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

	log.Printf("%v", res.Header.Get("Cache-Control"))
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

func (b Auth) Validate(ctx context.Context, token string) error {
	tok, err := jwt.ParseWithClaims(token, &Claims{}, func(tok *jwt.Token) (interface{}, error) {
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

	if err != nil {
		return fmt.Errorf("failed to parse jwt: %s", err)
	}

	claims := tok.Claims.(*Claims)
	clientID := os.Getenv("DS_GOOGLE_CLIENT_ID")

	// validates expiry information
	if err := claims.Valid(); err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// validates the token was signed with our client ID
	if !claims.VerifyAudience(clientID, true) {
		return errors.New("clientID in token did not match server")
	}

	// validates the token issuer matches expectations
	if !claims.VerifyIssuer(googleIssuer, true) {
		return errors.New("invalid issuer")
	}

	log.Printf("Token email: %s", claims.Email)
	return nil
}
