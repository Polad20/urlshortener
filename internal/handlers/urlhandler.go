package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Polad20/urlshortener/internal/shortener"
	"github.com/Polad20/urlshortener/internal/storage"
	"github.com/go-chi/chi/v5"
)

var (
	errNotUnique         = errors.New("not unique original url")
	errNotUniqueShortUrl = errors.New("not unique short url")
)

type Handler struct {
	*chi.Mux
	repo storage.Storage
}

func NewHandler(repo storage.Storage) *Handler {
	h := &Handler{
		Mux:  chi.NewMux(),
		repo: repo,
	}

	h.Post("/", h.saveURL())
	h.Get("/", h.getURL())

	return h
}

func (h *Handler) saveURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		service := shortener.NewShortener()
		if err := json.NewDecoder(r.Body).Decode(service); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		service.Shorten()
		shortened, err := h.repo.SaveData(service.OriginalURL, service.ShortURL)
		if err != nil {
			if err == errNotUniqueShortUrl {
				service.Reshorten(service.ShortURL)
				shortened = service.ShortURL
			}
			if err == errNotUnique {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		encoder.Encode(shortened)
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) getURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		service := shortener.NewShortener()
		if err := json.NewDecoder(r.Body).Decode(service); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		og, err := h.repo.GetData(service.ShortURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		encoder.Encode(og)
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
	}
}
