package http

import (
	"encoding/json"
	"net/http"

	"github.com/eriktate/skribe"
	log "github.com/sirupsen/logrus"
)

func HandleDoc(docStore skribe.DocStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		part := shiftPath(r)

		switch part {
		case "":
			switch r.Method {
			case http.MethodPost:
				var doc skribe.Doc
				if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
					log.Error(err)
					badRequest(w, "invalid request body, could not save document")
					return
				}

				if err := docStore.PutDoc(r.Context(), doc); err != nil {
					log.Error(err)
					serverError(w, "something went wrong while saving document")
					return
				}

				noContent(w)
				return
			default:
				notAllowed(w)
				return
			}
		case "list":
			path := shiftPath(r)
			switch r.Method {
			case http.MethodGet:
				docs, err := docStore.ListPath(r.Context(), path)
				if err != nil {
					log.Error(err)
					serverError(w, "something went wrong while listing documents")
					return
				}

				data, err := json.Marshal(docs)
				if err != nil {
					log.Error(err)
					serverError(w, "something went wrong while serializing documents")
					return
				}

				statusOk(w, data)
				return
			default:
				notAllowed(w)
				return
			}
		default:
			switch r.Method {
			case http.MethodGet:
				log.WithField("path", part).Info("getting doc")
				doc, err := docStore.GetDoc(r.Context(), part)
				if err != nil {
					if skribe.CheckDoesNotExist(err) {
						notFound(w)
						return
					}

					log.Error(err)
					serverError(w, "something went wrong while fetching document")
					return
				}

				data, err := json.Marshal(doc)
				if err != nil {
					log.Error(err)
					serverError(w, "something went wrong while serializing document")
					return
				}

				statusOk(w, data)
				return
			case http.MethodDelete:
				if err := docStore.RemoveDoc(r.Context(), part); err != nil {
					log.Error(err)
					serverError(w, "something went wrong while deleting document")
					return
				}

				noContent(w)
				return
			default:
				notAllowed(w)
				return
			}
		}
	})
}
