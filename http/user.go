package http

import (
	"encoding/json"
	"net/http"

	"github.com/docshelf/docshelf"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// ID is a struct for marshaling to and from JSON documents containing an ID.
type ID struct {
	ID string `json:"id"`
}

// A UserHandler has methods that can handle HTTP requests for Users.
type UserHandler struct {
	userStore docshelf.UserStore
	log       *logrus.Logger
}

// NewUserHandler returns a UserHandler struct using the given UserStore and Logger instance.
func NewUserHandler(userStore docshelf.UserStore, logger *logrus.Logger) UserHandler {
	return UserHandler{
		userStore: userStore,
		log:       logger,
	}
}

// GetCurrentUser handles requests for fetching the currently logged in user.
func (h UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := getContextUser(r.Context())
	if err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while fetching the current user")
		return
	}

	data, err := json.Marshal(user)
	if err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while serializing user")
		return
	}

	okJSON(w, data)
}

// GetUsers handles requests for listing all Users.
func (h UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userStore.ListUsers(r.Context())
	if err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while fetching user list")
		return
	}

	data, err := json.Marshal(users)
	if err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while serializing user list")
		return
	}

	okJSON(w, data)
}

// PostUser handles requests for posting new (or updating existing) Users.
func (h UserHandler) PostUser(w http.ResponseWriter, r *http.Request) {
	var user docshelf.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.log.Error(err)
		badRequest(w, "invalid request body, could not save user")
		return
	}

	// TODO (erik): Adjust the cost parameter once we can benchmark the time spent hashing the password.
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Token), 12)

	if err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while saving user information")
	}

	user.Token = string(hashed)
	id, err := h.userStore.PutUser(r.Context(), user)
	if err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while saving user information")
		return
	}

	data, err := json.Marshal(ID{id})
	if err != nil {
		h.log.Error(err)
		serverError(w, "user was saved, but the id couldn't be returned")
		return
	}

	okJSON(w, data)
}

// GetUser handles requests for fetching specific Users.
func (h UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.userStore.GetUser(r.Context(), id)
	if err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while fetching user")
		return
	}

	data, err := json.Marshal(user)
	if err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while serializing user")
		return
	}

	okJSON(w, data)
}

// DeleteUser handles requests for deleting specific Users.
func (h UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.userStore.RemoveUser(r.Context(), id); err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while deleting user")
	}

	noContent(w)
}
