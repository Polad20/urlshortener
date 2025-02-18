package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Polad20/urlshortener/internal/handlers"
	"github.com/Polad20/urlshortener/internal/middleware"
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
	switch storageType {
	case "in-memory":
		repo := inmem.NewInmem()
		newShortener := shortener.NewShortener(repo)
		r = handlers.NewHandler(repo, newShortener)
	case "postgres":
		repo, err := pg.NewPostgresStorage()
		if err != nil {
			log.Fatal("Ошибка создания нового экземпляра PostgresStorage")
		}
		newShortener := shortener.NewShortener(repo)
		r = handlers.NewHandler(repo, newShortener)
	}
	httpHandler := http.Handler(r)
	httpHandler = middleware.MiddlewareBrotliEncoder(httpHandler)
	log.Fatal(http.ListenAndServe(":8080", httpHandler))
}
