package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/docshelf/docshelf"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
)

// A Server is a collection of stores that get wired up to HTTP endpoint.
type Server struct {
	host string
	port uint
	log  *logrus.Logger

	DocHandler  DocHandler
	UserStore   docshelf.UserStore
	GroupStore  docshelf.GroupStore
	PolicyStore docshelf.PolicyStore
	Auth        docshelf.Authenticator
}

// NewServer returns a new Server struct.
func NewServer(host string, port uint, logger *logrus.Logger) Server {
	return Server{
		host: host,
		port: port,
		log:  logger,
	}
}

// Start fires up an HTTP server and listens for incoming requests.
func (s Server) Start() error {
	s.log.WithField("host", s.host).WithField("port", s.port).Info("server starting")
	// if err := s.CheckStores(); err != nil {
	// 	return err
	// }

	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.host, s.port), s.buildRoutes()); err != nil {
		return err
	}

	return nil
}

// CheckHandlers returns an error if the Server contains any invalid handlers.
func (s Server) CheckHandlers() error {
	if s.UserStore == nil {
		return errors.New("no UserStore set")
	}

	if s.GroupStore == nil {
		return errors.New("no GroupStore set")
	}

	if s.PolicyStore == nil {
		return errors.New("no PolicyStore set")
	}

	if s.Auth == nil {
		return errors.New("no Authenticator set")
	}

	return nil
}

func (s Server) buildRoutes() chi.Router {
	mux := chi.NewRouter()
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:1234"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	userHandler := NewUserHandler(s.UserStore, s.log)
	mux.Use(cors.Handler)
	mux.Route("/api", func(r chi.Router) {
		r.Use(Authentication(s.UserStore))
		r.Route("/user", func(r chi.Router) {
			r.Get("/", userHandler.GetUsers)
			r.Post("/", userHandler.PostUser)
			r.Get("/{id}", userHandler.GetUser)
			r.Delete("/{id}", userHandler.DeleteUser)
		})

		r.Route("/doc", func(r chi.Router) {
			r.Post("/", s.DocHandler.PostDoc)
			r.Get("/list", s.DocHandler.GetList)
			r.Get("/{path}", s.DocHandler.GetDoc)
			r.Delete("/{path}", s.DocHandler.DeleteDoc)
		})

		r.Post("/tag", s.DocHandler.PostTag)
	})

	mux.Get("/doc/{path}", s.DocHandler.RenderDoc)
	mux.Post("/login", s.handleLogin)

	mux.Handle("/*", http.FileServer(http.Dir("./ui/dist/")))

	return mux
}

func (s Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var login docshelf.User

	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		s.log.Error(err)
		badRequest(w, "invalid authentication data")
		return
	}

	if err := s.Auth.Authenticate(r.Context(), login.Email, login.Token); err != nil {
		s.log.Error(err)
		unauthorized(w, "invalid credentials")
		return
	}

	user, err := s.UserStore.GetUser(r.Context(), login.Email)
	if err != nil {
		s.log.Error(err)
		serverError(w, "something went wrong while verifying credentials")
		return
	}

	// TODO (erik): Need to sign this data and add an expiration.
	// Also may need to expand the data stored and remove the HttpOnly.
	identity := http.Cookie{
		Name:     "session",
		Value:    user.ID,
		HttpOnly: true,
	}

	http.SetCookie(w, &identity)
	noContent(w)
}

// everything down here is setup for attaching certain data to the request context.
type contextKey string

const userKey = contextKey("ds-user")

func getContextUser(ctx context.Context) (docshelf.User, error) {
	if user, ok := ctx.Value(userKey).(docshelf.User); ok {
		return user, nil
	}

	return docshelf.User{}, errors.New("no user found in context")
}
