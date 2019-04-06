package http

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func serverError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := w.Write([]byte(msg)); err != nil {
		log.WithError(err).Error()
	}
}

func okJson(w http.ResponseWriter, data []byte) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.WithError(err).Error()
	}
}

func okHtml(w http.ResponseWriter, data []byte) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
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
