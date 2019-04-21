package http

import (
	"context"
	"net/http"

	"github.com/docshelf/docshelf"
)

// Authentication is a middleware that verifies a user's identity before passing control
// to the underlying HTTP handler. If the user is verified, their data is pulled and attached
// to the request context so it can be referenced at all stages later on.
func Authentication(userStore docshelf.UserStore) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := r.Cookie("session")
			if err != nil {
				unauthorized(w, "no active session")
				return
			}

			user, err := userStore.GetUser(r.Context(), session.Value)
			if err != nil {
				return
			}

			// attach user struct to the context before passing to the next handler
			ctx := context.WithValue(r.Context(), userKey, user)
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
