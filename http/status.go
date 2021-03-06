package http

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func okJSON(w http.ResponseWriter, data []byte) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.WithError(err).Error()
	}
}

func okHTML(w http.ResponseWriter, data []byte) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.WithError(err).Error()
	}
}

func noContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
	if _, err := w.Write(nil); err != nil {
		log.WithError(err).Error()
	}
}

// NOTE: This isn't a typical 3xx redirect. It's a 200 with the refresh header
// set.
func redirect(w http.ResponseWriter, url string) {
	w.Header().Add("Refresh", fmt.Sprintf("0;url=%s", url))
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(nil); err != nil {
		log.WithError(err).Error()
	}
}

func badRequest(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	if _, err := w.Write([]byte(msg)); err != nil {
		log.WithError(err).Error()
	}
}

func notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	if _, err := w.Write(nil); err != nil {
		log.WithError(err).Error()
	}
}

func unauthorized(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusUnauthorized)
	if _, err := w.Write([]byte(msg)); err != nil {
		log.WithError(err).Error()
	}
}

func serverError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := w.Write([]byte(msg)); err != nil {
		log.WithError(err).Error()
	}
}
