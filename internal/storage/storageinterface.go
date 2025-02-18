package storage

import (
	"github.com/Polad20/urlshortener/internal/model"
)

type Storage interface {
	SaveURL(userID, shortURL, originalURL string) error
	GetURLsByUser(userID string) ([]model.ShortenedURL, error)
}
