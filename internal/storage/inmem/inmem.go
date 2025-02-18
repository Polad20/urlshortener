package storage

import (
	"context"
	"sync"

	"github.com/Polad20/urlshortener/internal/model"
)

type Inmem struct {
	urlList map[string][]model.ShortenedURL
	lock    sync.Mutex
}

func NewInmem() *Inmem {
	memstor := &Inmem{}
	memstor.urlList = make(map[string][]model.ShortenedURL)
	return memstor
}

func (storage *Inmem) SaveURL(userID, shortURL, originalURL string) error {
	storage.lock.Lock()
	defer storage.lock.Unlock()
	shortenedURL := model.ShortenedURL{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	if _, ok := storage.urlList[userID]; !ok {
		storage.urlList[userID] = []model.ShortenedURL{}
	}
	storage.urlList[userID] = append(storage.urlList[userID], shortenedURL)
	return nil
}

func (storage *Inmem) GetURLsByUser(userID string) ([]model.ShortenedURL, error) {
	storage.lock.Lock()
	defer storage.lock.Unlock()
	urls, ok := storage.urlList[userID]
	if !ok {
		return nil, nil
	}
	return urls, nil
}

func (storage *Inmem) Ping(ctx context.Context) error {
	return nil
}
