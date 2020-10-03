package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/docshelf/docshelf"
)

const tokenURL = "https://github.com/login/oauth/access_token" // nolint:gosec
const userURL = "https://api.github.com/user"
const emailURL = "https://api.github.com/user/emails"

// Github config for authenticating through oauth.
type Github struct {
	ClientID     string
	ClientSecret string

	userStore docshelf.UserStore
	client    *http.Client
}

// NewGithub creates a new Github authenticator with the given clientID and
// clientSecret.
func NewGithub(userStore docshelf.UserStore, clientID, clientSecret string) Github {
	return Github{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		userStore:    userStore,
		client:       http.DefaultClient,
	}
}

type oauthResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type userResponse struct {
	Name string `json:"name"` // we only need the name from the user object
}

type emailResponse struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

func (g Github) getUser(accessToken string) (userResponse, error) {
	req, err := http.NewRequest(http.MethodGet, userURL, nil)
	if err != nil {
		return userResponse{}, fmt.Errorf("failed to create user request: %w", err)
	}

	req.Header.Add("Authorization", "token "+accessToken)
	res, err := g.client.Do(req)
	if err != nil {
		return userResponse{}, fmt.Errorf("failed to send user request: %w", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return userResponse{}, fmt.Errorf("failed to read user response: %w", err)
	}

	var user userResponse
	if err := json.Unmarshal(body, &user); err != nil {
		return user, fmt.Errorf("failed to unmarshal user response: %w", err)
	}

	return user, nil
}

func (g Github) getPrimaryEmail(accessToken string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, emailURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create email request: %w", err)
	}

	req.Header.Add("Authorization", "token "+accessToken)
	res, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send email request: %w", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", fmt.Errorf("failed to read email response: %w", err)
	}

	var emails []emailResponse
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", fmt.Errorf("failed to unmarshal email response: %w", err)
	}

	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}

	return "", errors.New("no primary email found")

}

// Authenticate implements the dochself.Authenticator interface. It takes
// a temporary auth code and uses it to request an access token from github.
// If authorization is successful, the primary email for the authorized user
// is used to login or. The email argument is completely ignored, but required
// to implement the docshelf.Authenticator interface.
func (g Github) Authenticate(ctx context.Context, email, token string) (docshelf.User, error) {
	req, err := http.NewRequest(http.MethodPost, tokenURL, nil)
	if err != nil {
		return docshelf.User{}, fmt.Errorf("failed to create token request: %w", err)
	}

	query := req.URL.Query()
	query.Add("client_id", g.ClientID)
	query.Add("client_secret", g.ClientSecret)
	query.Add("code", token)
	req.URL.RawQuery = query.Encode()

	req.Header.Set("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return docshelf.User{}, fmt.Errorf("failed to get access token: %w", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return docshelf.User{}, fmt.Errorf("failed to read access token: %w", err)
	}

	var oRes oauthResponse
	if err := json.Unmarshal(body, &oRes); err != nil {
		return docshelf.User{}, fmt.Errorf("failed to unmarshal access token: %w", err)
	}

	user, err := g.getUser(oRes.AccessToken)
	if err != nil {
		return docshelf.User{}, err
	}

	primaryEmail, err := g.getPrimaryEmail(oRes.AccessToken)
	if err != nil {
		return docshelf.User{}, err
	}

	return getOrPutUser(ctx, g.userStore, primaryEmail, user.Name)
}
