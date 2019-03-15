package http

import (
	"encoding/json"
	"net/http"

	"github.com/eriktate/skribe"
	log "github.com/sirupsen/logrus"
)

type ID struct {
	ID string `json:"id"`
}

func HandleUser(userStore skribe.UserStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		part := shiftPath(r)
		log.WithField("part", part).Info("Path")

		if part == "" {
			switch r.Method {
			case http.MethodGet:
				users, err := userStore.ListUsers(r.Context())
				if err != nil {
					log.Error(err)
					serverError(w, "something went wrong while fetching user list")
					return
				}

				data, err := json.Marshal(users)
				if err != nil {
					log.Error(err)
					serverError(w, "something went wrong while serializing user list")
					return
				}

				okJson(w, data)
				return

			case http.MethodPost:
				var user skribe.User
				if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
					log.Error(err)
					badRequest(w, "invalid request body, could not create user")
					return
				}

				id, err := userStore.PutUser(r.Context(), user)
				if err != nil {
					log.Error(err)
					serverError(w, "something went wrong while saving user")
					return
				}

				data, err := json.Marshal(ID{id})
				if err != nil {
					log.Error(err)
					serverError(w, "user was created, but the id couldn't be returned")
					return
				}

				okJson(w, data)
				return

			default:
				badRequest(w, "method unsupported")
				return
			}
		}

		switch r.Method {
		case http.MethodGet:
			user, err := userStore.GetUser(r.Context(), part)
			if err != nil {
				log.Error(err)
				serverError(w, "something went wrong while fetching user")
				return
			}

			data, err := json.Marshal(user)
			if err != nil {
				log.Error(err)
				serverError(w, "something went wrong while serializing user")
				return
			}

			okJson(w, data)
			return
		case http.MethodDelete:
			if err := userStore.RemoveUser(r.Context(), part); err != nil {
				log.Error(err)
				serverError(w, "something went wrong while deleting user")
			}

			noContent(w)
			return
		default:
			badRequest(w, "method unsupported")
			return
		}
	})
}
