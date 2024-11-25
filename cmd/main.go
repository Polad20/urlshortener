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
		log.Fatal("Error loading .env file, using environment variables or defaults")
	}
	/* serverAdressDefault := getEnvOrDefault("SERVER_ADRESS", "localhost:8080")
	 baseURLDefault := getEnvOrDefault("BASE_URL", "http://localhost:8080")
	 fileStorageDefault := getEnvOrDefault("FILE_STORAGE_PATH", "")

	 serverAdвress := flag.String("a", serverAdressDefault, "Server Adress")
	 baseURL := flag.String("b", baseURLDefault, "Base URL")
	 fileStoragePath := flag.String("f", fileStorageDefault, "File Storage Path")

	flag.Parse()
	*/
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

func getEnvOrDefault(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
