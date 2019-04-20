package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/docshelf/docshelf"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
)

// A Server is a collection of stores that get wired up to HTTP endpoint.
type Server struct {
	host string
	port uint
	log  *logrus.Logger

	DocHandler  DocHandler
	UserHandler UserHandler
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
	mux.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Get("/", s.UserHandler.GetUsers)
			r.Post("/", s.UserHandler.PostUser)
			r.Get("/{id}", s.UserHandler.GetUser)
			r.Delete("/{id}", s.UserHandler.DeleteUser)
		})

		r.Route("/doc", func(r chi.Router) {
			r.Post("/", s.DocHandler.PostDoc)
			r.Get("/list", s.DocHandler.GetList)
			r.Get("/list/{prefix}", s.DocHandler.GetList)
			r.Get("/{path}", s.DocHandler.GetDoc)
			r.Delete("/{path}", s.DocHandler.DeleteDoc)
		})

		r.Post("/tag", s.DocHandler.PostTag)
	})

	mux.Get("/doc/{path}", s.DocHandler.RenderDoc)

	return mux
}
