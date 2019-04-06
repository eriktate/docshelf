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
		HandleDoc(s.DocStore).ServeHTTP(w, r)
	case "group":
		statusOk(w, []byte("unimplemented"))
	case "policy":
		statusOk(w, []byte("unimplemented"))
	}
}

func shiftPath(r *http.Request) string {
	log.WithField("path", r.URL.Path).Info("Path")
	if r.URL.Path == "/" {
		return ""
	}

	parts := strings.Split(r.URL.Path, "/")
	newPath := "/"
	if len(parts) > 2 {
		newPath += strings.Join(parts[2:], "/")
	}

	r.URL.Path = newPath

	return parts[1]
}

func serverError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := w.Write([]byte(msg)); err != nil {
		log.WithError(err).Error()
	}
}

func statusOk(w http.ResponseWriter, data []byte) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.WithError(err).Error()
	}
}

func badRequest(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	if _, err := w.Write([]byte(msg)); err != nil {
		log.WithError(err).Error()
	}
}

func noContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
	if _, err := w.Write(nil); err != nil {
		log.WithError(err).Error()
	}
}

func notAllowed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	if _, err := w.Write(nil); err != nil {
		log.WithError(err).Error()
	}
}

func notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	if _, err := w.Write(nil); err != nil {
		log.WithError(err).Error()
	}
}
