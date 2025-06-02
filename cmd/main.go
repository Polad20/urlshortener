package main

import (
	"net/http"

	"github.com/Polad20/urlshortener/config"
	"github.com/Polad20/urlshortener/internal/auth"
	"github.com/Polad20/urlshortener/internal/handlers"
	"github.com/Polad20/urlshortener/internal/shortener"
	inmem "github.com/Polad20/urlshortener/internal/storage/inmem"
	pg "github.com/Polad20/urlshortener/internal/storage/pg"
	"github.com/Polad20/urlshortener/logs"
	log "github.com/rs/zerolog/log"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatal().Err(err).Msgf("Error loading Config: %v", err)
	}

	var r *handlers.Handler
	logs.InitLogger(config.AppConfig.Log)
	storageType := config.AppConfig.Repository.Type
	authKey := config.AppConfig.Auth.Key
	if authKey == "" {
		log.Fatal().Msg("AUTH_SECRET_KEY environment variable not set for authentication middleware")
	}
	authKeyBytes := []byte(authKey)
	authMiddleware := auth.New(authKeyBytes)
	newShortener := shortener.NewShortener()
	switch storageType {
	case "in-memory":
		repo := inmem.NewInmem()
		r = handlers.NewHandler(repo, newShortener, authMiddleware)
	case "postgres":
		repo, err := pg.NewPostgresStorage()
		if err != nil {
			log.Fatal().Msg("Ошибка создания нового экземпляра PostgresStorage")
		}
		r = handlers.NewHandler(repo, newShortener, authMiddleware)
	}
	log.Fatal().Err(http.ListenAndServe(":8080", r)).Msg("Сервер остановлен.")
}
