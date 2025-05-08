package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Polad20/urlshortener/internal/auth"
	"github.com/Polad20/urlshortener/internal/handlers"
	"github.com/Polad20/urlshortener/internal/shortener"
	inmem "github.com/Polad20/urlshortener/internal/storage/inmem"
	pg "github.com/Polad20/urlshortener/internal/storage/pg"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file, using environment variables or defaults")
	}

	var r *handlers.Handler

	storageType := os.Getenv("REPO")
	authKey := os.Getenv("KEY")
	if authKey == "" {
		log.Fatal("AUTH_SECRET_KEY environment variable not set for authentication middleware")
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
			log.Fatal("Ошибка создания нового экземпляра PostgresStorage")
		}
		r = handlers.NewHandler(repo, newShortener, authMiddleware)
	}
	log.Fatal(http.ListenAndServe(":8080", r))
}
