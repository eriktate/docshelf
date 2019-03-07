package http

import (
	"fmt"
	"net/http"

	"github.com/eriktate/skribe"
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
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.addr, s.port), http.HandlerFunc(handler)); err != nil {
		return err
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}
