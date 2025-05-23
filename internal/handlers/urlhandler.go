package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Polad20/urlshortener/internal/auth"
	"github.com/Polad20/urlshortener/internal/middleware"
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

func NewHandler(repo storage.Storage, shortener *shortener.Shortener, authMiddleware *auth.Auth) *Handler {
	h := &Handler{
		Mux:       chi.NewMux(),
		repo:      repo,
		shortener: shortener,
	}
	h.Use(authMiddleware.MiddlewareAuth)
	h.Use(middleware.MiddlewareBrotliEncoder)
	h.Get("/{id}", h.RedirectHandler())
	h.Post("/api/pg/shorten/batch", h.SaveBaseURL())
	h.Post("/api/user/urls", h.deleteBatch())
	h.Post("/api/inmem/shorten", h.saveURL())
	h.Get("/api/inmem/user/urls", h.getURL())
	h.Get("/api/pg/ping", h.pingHandler())

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
			log.Println("Can`t convert userID to string")
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

func (h *Handler) RedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID")
		userIDValue, ok := userID.(string)
		if !ok {
			http.Error(w, "Error: userID is not a string", http.StatusBadRequest)
			log.Printf("userID is not a string")
			return
		}
		id := chi.URLParam(r, "id")
		if id == "" {
			log.Printf("Redirect Handler Error: ID not found in URL path")
			http.Error(w, "Invalid request path", http.StatusBadRequest)
			return
		}
		shortURL := fmt.Sprintf("http://localhost:8080/%s", id)
		originalURL, err := h.repo.FindUsersOrigURL(userIDValue, shortURL)
		if err != nil {
			log.Printf("Redirect Handler Error: Failed 	to get original URL for '%s': %v", shortURL, err)
			http.Error(w, "Cant find original url for given short", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
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
			log.Printf("Error decoding body: %v", err)
			return
		}
		for _, i := range memory {
			userIDvalue, ok := userID.(string)
			if !ok {
				http.Error(w, "userID in context not a string", http.StatusInternalServerError)
				log.Printf("Error: userID in context not a string, %v", userID)
				return
			}
			newDBentry, err := pg.DbSavePrepare(userIDvalue, i, h.shortener)
			if err != nil {
				http.Error(w, "Error preparing DB entry ", http.StatusInternalServerError)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(clientResponses); err != nil {
			log.Printf("Error encoding batch response: %v", err)
		}
	}
}

func (h *Handler) deleteBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDinter := r.Context().Value("userID")
		if userIDinter == nil {
			http.Error(w, "You don`t have userID to delete URL`s", http.StatusBadRequest)
			return
		}
		userID, ok := userIDinter.(string)
		if !ok {
			http.Error(w, "Can`t convert userID to string", http.StatusInternalServerError)
			return
		}
		var incoming []string
		ch := make(chan string)
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		err := decoder.Decode(&incoming)
		if err != nil {
			http.Error(w, "Error decoding body", http.StatusBadRequest)
			return
		}
		go func() {
			defer close(ch)
			for _, i := range incoming {
				ch <- i
			}
		}()
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PANIC in async batch delete for user %s: %v", userID, r)
				}
			}()
			pgrepo, ok := h.repo.(*pg.PostgresStorage)
			if !ok {
				log.Printf("Error: storage backend is not PostgreSQL,batch delete requires PostgreSQL")
				return
			}
			tx, err := pgrepo.DB.BeginTx(context.Background(), nil)
			if err != nil {
				log.Printf("Error: storage backend is not PostgreSQL,batch delete requires PostgreSQL")
				return
			}
			defer tx.Rollback()
			const maxBatchsize = 500
			currentBatchSlice := make([]string, 0, maxBatchsize)
			for ach := range ch {
				currentBatchSlice = append(currentBatchSlice, ach)
				if len(currentBatchSlice) == maxBatchsize {
					err := pgrepo.DeleteURLs(userID, currentBatchSlice, tx)
					if err != nil {
						log.Printf("Batch Error - failed to call method DeleteURLs")
						return
					}
					currentBatchSlice = make([]string, 0, maxBatchsize)
				}
			}
			if len(currentBatchSlice) > 0 {
				err := pgrepo.DeleteURLs(userID, currentBatchSlice, tx)
				if err != nil {
					log.Printf("Batch Error - failed for final delete")
					return
				}
			}
			err = tx.Commit()
			if err != nil {
				log.Printf("Batch Error - failed to commit transaction")
				return
			}
			log.Printf("Batch delete completed succesfully")

		}()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
	}
}
