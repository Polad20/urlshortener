package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Polad20/urlshortener/internal/shortener"
	"github.com/Polad20/urlshortener/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	*chi.Mux
	repo      storage.Storage
	shortener *shortener.Shortener
}

func NewHandler(repo storage.Storage) *Handler {
	h := &Handler{
		Mux:  chi.NewMux(),
		repo: repo,
	}

	h.Post("/", h.saveURL())
	h.Get("/api/user/urls", h.getURL())
	h.Get("/ping", h.pingHandler())

	return h
}

func (h *Handler) saveURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDValue := r.Context().Value("userID")
		if userIDValue == nil {
			log.Println("Can`t get userID from Context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		userID, ok := userIDValue.(string)
		if !ok {
			log.Println("Can`t convert userID to string")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var req struct {
			OriginalURL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		shortURL, err := h.shortener.Shorten(userID, req.OriginalURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		encoder.Encode(map[string]string{"result": shortURL})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) getURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDinter := r.Context().Value("userID")
		if userIDinter == nil {
			http.Error(w, "You don`t have userID to get URL`s", http.StatusBadRequest)
			return
		}
		userID, ok := userIDinter.(string)
		if !ok {
			log.Println("Can`t conver userID to string")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		urls, err := h.repo.GetURLsByUser(userID)
		if err != nil {
			http.Error(w, "Error getting url`s", http.StatusInternalServerError)
			return
		}
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		encoder.Encode(urls)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

	}
}

func (h *Handler) pingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.repo.Ping(context.Background())
		if err != nil {
			http.Error(w, "DB connection error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
