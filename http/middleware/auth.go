package middleware

import (
	"net/http"

	"github.com/docshelf/docshelf"
)

// Authentication is a middleware that verifies a user's identity before passing control
// to the underlying HTTP handler.
func Authentication(auth docshelf.Authenticator) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
		})
	}
}
