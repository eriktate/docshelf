package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/eriktate/skribe"
	log "github.com/sirupsen/logrus"
)

// A Server is a collection of stores that get wired up to HTTP endpoint.
type Server struct {
	DocStore    skribe.DocStore
	UserStore   skribe.UserStore
	GroupStore  skribe.GroupStore
	PolicyStore skribe.PolicyStore
	Auth        skribe.Authenticator

	addr string
	port uint
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
	// if err := s.CheckStores(); err != nil {
	// 	return err
	// }

	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.addr, s.port), http.HandlerFunc(s.handler)); err != nil {
		return err
	}

	return nil
}

// CheckStores returns an error if the Server is missing any required Stores.
func (s Server) CheckStores() error {
	if s.DocStore == nil {
		return errors.New("no DocStore set")
	}

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

func (s Server) handler(w http.ResponseWriter, r *http.Request) {
	part := shiftPath(r)
	if part == "" {
		log.Error(errors.New("no resource specified"))
		badRequest(w, "you need to specify a resource")
		return
	}

	switch part {
	case "user":
		HandleUser(s.UserStore).ServeHTTP(w, r)
		return
	case "doc":
		w.Write([]byte("unimplemented"))
	case "group":
		w.Write([]byte("unimplemented"))
	case "policy":
		w.Write([]byte("unimplemented"))
	}
}

func shiftPath(r *http.Request) string {
	log.WithField("path", r.URL.Path).Info("Path")
	if r.URL.Path == "/" {
		return ""
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 2 {
		r.URL.Path = "/" + strings.Join(parts[2:len(parts)], "/")
	}

	return parts[1]
}

func serverError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}

func statusOk(w http.ResponseWriter, data []byte) {
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func badRequest(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(msg))
}

func noContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
	w.Header().Add("Content-Type", "application/json")
	w.Write(nil)
}
