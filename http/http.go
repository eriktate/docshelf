package http

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"

	"github.com/eriktate/skribe"
	"github.com/russross/blackfriday"
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
	log.Info("Starting doc server...")
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

	switch part {
	case "api":
		s.handleAPI(w, r)
		return
	case "doc":
		s.handleDoc(w, r)
	default:
		w.Write([]byte("unimplemented"))
	}
}

func (s Server) handleAPI(w http.ResponseWriter, r *http.Request) {
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
	default:
		notFound(w)
	}

}

func (s Server) handleDoc(w http.ResponseWriter, r *http.Request) {
	doc, err := s.DocStore.GetDoc(r.Context(), r.URL.Path[1:len(r.URL.Path)])
	if err != nil {
		log.Error(err)
		serverError(w, "could not render page")
		return
	}

	dom := blackfriday.Run(doc.Content)
	doc.Content = dom

	// TODO (erik): Need to embed this template in the binary rather than reading off of
	// the file system.
	f, err := ioutil.ReadFile("./template.html")
	if err != nil {
		log.Error(err)
		serverError(w, "could not render page")
		return
	}

	tmpl, err := template.New("doc").Parse(string(f))
	if err != nil {
		log.Error(err)
		serverError(w, "could not render page")
		return
	}

	data := make([]byte, 0)
	output := bytes.NewBuffer(data)
	if err := tmpl.Execute(output, doc); err != nil {
		log.Error(err)
		serverError(w, "could not render page")
		return
	}

	okHtml(w, output.Bytes())
}

func peekPath(r *http.Request) string {
	if r.URL.Path == "/" {
		return ""
	}

	return strings.Split(r.URL.Path, "/")[1]
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
