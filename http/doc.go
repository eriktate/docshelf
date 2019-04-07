package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"

	"github.com/eriktate/docshelf"
	"github.com/go-chi/chi"
	"github.com/russross/blackfriday"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// A TagReq is a request to apply tags to a given document.
type TagReq struct {
	Path string
	Tags []string
}

// A DocHandler has methods that can handle HTTP requests for Docs.
type DocHandler struct {
	docStore docshelf.DocStore
	log      *logrus.Logger
}

// NewDocHandler returns a DocHandler struct using the given DocStore and Logger instance.
func NewDocHandler(docStore docshelf.DocStore, logger *logrus.Logger) DocHandler {
	return DocHandler{
		docStore: docStore,
		log:      logger,
	}
}

// PostDoc handles requests for posting new (or existing) Docs.
func (h DocHandler) PostDoc(w http.ResponseWriter, r *http.Request) {
	var doc docshelf.Doc
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		h.log.Error(err)
		badRequest(w, "invalid request body, could not save document")
		return
	}

	if err := h.docStore.PutDoc(r.Context(), doc); err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while saving document")
		return
	}

	noContent(w)
}

// GetList handles requests for listing Docs by path prefix.
func (h DocHandler) GetList(w http.ResponseWriter, r *http.Request) {
	prefix := chi.URLParam(r, "prefix")
	tags := strings.Split(r.URL.Query().Get("tags"), ",")
	if len(tags) == 1 && tags[0] == "" {
		tags = nil
	}
	h.log.WithField("tags", tags).WithField("prefix", prefix).Info("Getting list")
	docs, err := h.docStore.ListDocs(r.Context(), prefix, tags...)
	if err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while listing documents")
		return
	}

	data, err := json.Marshal(docs)
	if err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while serializing documents")
		return
	}

	okJSON(w, data)
}

// PostTag handles requests for posting tags to an existing Doc.
func (h DocHandler) PostTag(w http.ResponseWriter, r *http.Request) {
	var req TagReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error(err)
		badRequest(w, "invalid format for tagging documents")
		return
	}

	if err := h.docStore.TagDoc(r.Context(), req.Path, req.Tags...); err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while tagging document")
		return
	}

	noContent(w)
}

// GetDoc handles requests for fetching specific Docs.
func (h DocHandler) GetDoc(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "path")
	doc, err := h.docStore.GetDoc(r.Context(), path)
	if err != nil {
		if docshelf.CheckDoesNotExist(err) {
			notFound(w)
			return
		}

		h.log.Error(err)
		serverError(w, "something went wrong while fetching document")
		return
	}

	data, err := json.Marshal(doc)
	if err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while serializing document")
		return
	}

	okJSON(w, data)
}

// DeleteDoc handles requests for removing specific Docs.
func (h DocHandler) DeleteDoc(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "path")
	if err := h.docStore.RemoveDoc(r.Context(), path); err != nil {
		h.log.Error(err)
		serverError(w, "something went wrong while deleting document")
		return
	}

	noContent(w)
}

// RenderDoc handles requests for rendering Documents as HTML.
func (h DocHandler) RenderDoc(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "path")

	doc, err := h.docStore.GetDoc(r.Context(), path)
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

	okHTML(w, output.Bytes())
}
