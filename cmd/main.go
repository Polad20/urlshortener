package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Polad20/urlshortener/internal/handlers"
	inmem "github.com/Polad20/urlshortener/internal/storage/inmem"
	pg "github.com/Polad20/urlshortener/internal/storage/pg"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading environment")
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

	log.Fatal(http.ListenAndServe(":8080", r))
}
