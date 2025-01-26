package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Polad20/urlshortener/internal/handlers"
	"github.com/Polad20/urlshortener/internal/middleware"
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
		r = handlers.NewHandler(repo)
	case "postgres":
		repo := pg.NewPostgres()
		r = handlers.NewHandler(repo)
	}
	httpHandler := http.Handler(r)
	httpHandler = middleware.MiddlewareBrotliEncoder(httpHandler)

	log.Fatal(http.ListenAndServe(":8080", httpHandler))
}
