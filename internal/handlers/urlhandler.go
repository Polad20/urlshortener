package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Polad20/urlshortener/internal/model"
	"github.com/Polad20/urlshortener/internal/shortener"
	"github.com/Polad20/urlshortener/internal/storage"
	"github.com/Polad20/urlshortener/internal/storage/pg"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	*chi.Mux
	repo      storage.Storage
	shortener *shortener.Shortener
}

func NewHandler(repo storage.Storage, shortener *shortener.Shortener) *Handler {
	h := &Handler{
		Mux:       chi.NewMux(),
		repo:      repo,
		shortener: shortener,
	}

	h.Post("/api/pg/shorten/batch", h.SaveBaseURL())
	h.Post("/api/inmem/shorten", h.saveURL())
	h.Get("/api/inmem/user/urls", h.getURL())
	h.Get("api/pg/ping", h.pingHandler())

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
		shortURL := h.shortener.Shorten()
		err := h.repo.SaveURL(userID, shortURL, req.OriginalURL)
		if err != nil {
			http.Error(w, "Failed to Save URL", http.StatusInternalServerError)
			log.Printf("Error saving URL to storage: %v", err)
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
		w.WriteHeader(http.StatusCreated)

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

func (h *Handler) SaveBaseURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID")
		var memory []model.Incoming
		var newDBvar []model.DbSave
		var clientResponses []model.ClientResponse
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		err := decoder.Decode(&memory)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Got some error decoding body: %v", err)
			return
		}
		for _, i := range memory {
			userIDvalue, ok := userID.(string)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("Error: userID in context not a string, %v", userID)
				return
			}
			newDBentry, err := pg.DbSavePrepare(userIDvalue, i, h.shortener)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("Error preparing DB entry: %v", err)
				return
			}
			newDBvar = append(newDBvar, newDBentry)
			clientRespSingle := model.ClientResponse{
				Correlation_id: newDBentry.Correlation_id,
				Short_url:      newDBentry.Short_url,
			}
			clientResponses = append(clientResponses, clientRespSingle)
		}
		pgRepo, ok := h.repo.(*pg.PostgresStorage)
		if !ok {
			http.Error(w, "Internal server error: storage type mismatch", http.StatusInternalServerError)
			log.Printf("Error: storage backend is not PostgreSQL, batch save requires PostgreSQL")
			return
		}
		err = pgRepo.BaseSave(r.Context(), newDBvar)
		if err != nil {
			http.Error(w, "Internal server error during database operation", http.StatusInternalServerError)
			log.Printf("Error saving batch to DB: %v", err)
			return
		}
		w.Header().Set("Content-Type", "apllication/json")
		w.WriteHeader(http.StatusCreated)
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(clientResponses); err != nil {
			log.Printf("Error encoding batch response: %v", err)
		}
	}
}
