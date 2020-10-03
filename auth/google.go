package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/docshelf/docshelf"
)

const googleKeyEndpoint = "https://www.googleapis.com/oauth2/v1/certs"
const googleIssuer = "accounts.google.com"

// Claims represent claimed data embedded in an ouath access token.
type Claims struct {
	jwt.StandardClaims

	Email string `json:"email"`
	Name  string `json:"name"`
}

// Google config for oauth authentication.
type Google struct {
	clientID     string
	clientSecret string
	userStore    docshelf.UserStore
	client       *http.Client
}

// NewGoogle returns a new Google oauth config.
func NewGoogle(userStore docshelf.UserStore, clientID, clientSecret string) Google {
	return Google{
		clientID:     clientID,
		clientSecret: clientSecret,
		userStore:    userStore,
		client:       http.DefaultClient,
	}
}

func (g Google) getGooglePublicKeys() (map[string]string, error) {
	res, err := g.client.Get(googleKeyEndpoint)
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

func (g Google) validate(ctx context.Context, token string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(token, &Claims{}, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
		}

		keys, err := g.getGooglePublicKeys()
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
		return nil, fmt.Errorf("failed to parse jwt: %s", err)
	}

	claims := tok.Claims.(*Claims)

	// validates expiry information
	if err := claims.Valid(); err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// validates the token was signed with our client ID
	if !claims.VerifyAudience(g.clientID, true) {
		return nil, errors.New("clientID in token did not match server")
	}

	// validates the token issuer matches expectations
	if !claims.VerifyIssuer(googleIssuer, true) {
		return nil, errors.New("invalid issuer")
	}

	return claims, nil
}

func (g Google) Authenticate(ctx context.Context, email, token string) (docshelf.User, error) {
	claims, err := g.validate(ctx, token)
	if err != nil {
		return docshelf.User{}, err
	}

	return getOrPutUser(ctx, g.userStore, claims.Email, claims.Name)
}
