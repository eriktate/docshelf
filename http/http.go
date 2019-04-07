package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/eriktate/skribe"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// A Server is a collection of stores that get wired up to HTTP endpoint.
type Server struct {
	DocHandler  DocHandler
	UserHandler UserHandler
	GroupStore  skribe.GroupStore
	PolicyStore skribe.PolicyStore
	Auth        skribe.Authenticator

	addr   string
	port   uint
	router chi.Router
}

// NewServer returns a new Server struct.
func NewServer(addr string, port uint) Server {
	return Server{
		addr: addr,
		port: port,
	}
}

// Start fires up an HTTP server and listens for incoming requests.
func (s Server) Start() error {
	log.Info("Starting doc server...")
	// if err := s.CheckStores(); err != nil {
	// 	return err
	// }

	// init API routes. Need to do this here instead of NewServer to make sure all handlers
	// are set properly.
	s.buildRoutes()

	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.addr, s.port), s.router); err != nil {
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

func (s Server) buildRoutes() {
	mux := chi.NewRouter()
	mux.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Get("", s.UserHandler.GetUsers)
			r.Post("", s.UserHandler.PostUser)
			r.Get("/{id}", s.UserHandler.GetUser)
			r.Delete("/{id}", s.UserHandler.DeleteUser)
		})

		mux.Route("/doc", func(r chi.Router) {
			r.Post("", s.DocHandler.PostDoc)
			r.Get("/list", s.DocHandler.GetList)
			r.Post("/tag", s.DocHandler.PostTag)
			r.Get("/{path}", s.DocHandler.GetDoc)
			r.Delete("/{path}", s.DocHandler.DeleteDoc)
		})
	})

	mux.Get("/doc/{path}", s.DocHandler.RenderDoc)

	s.router = mux
}
