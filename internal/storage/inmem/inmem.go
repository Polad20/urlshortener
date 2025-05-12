package inmem

import (
	"context"
	"fmt"
	"log"
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
	log.Printf("Saving URL for user %s: shortURL='%s', originalURL='%s'", userID, shortURL, originalURL)
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

func (storage *Inmem) FindUsersOrigURL(userID, shortURL string) (string, error) {
	pairsURL, ok := storage.urlList[userID]
	if !ok {
		log.Printf("Can`t find this Users URL`s")
		return "", fmt.Errorf("User Not Found")
	}
	for _, v := range pairsURL {
		if v.ShortURL == shortURL {
			return v.OriginalURL, nil
		}
	}
	return "", fmt.Errorf("Can`t find %s for current user", shortURL)
}
