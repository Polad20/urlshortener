package shortener

import (
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/Polad20/urlshortener/internal/storage"
)

type Shortener struct {
	OriginalURL string `json:"originalurl"`
	ShortURL    string `json:"shorturl"`
	randomizer  *rand.Rand
	charset     string
	urlLen      string
	myDomain    string
	storage     storage.Storage
}

// Creates new instance of Shortener service.
// Uses env variables to fill inner fields.
func NewShortener(storage storage.Storage) *Shortener {
	return &Shortener{
		randomizer: rand.New(rand.NewSource(time.Now().UnixNano())),
		charset:    os.Getenv("CHARSET"),
		urlLen:     os.Getenv("LENGTH"),
		myDomain:   os.Getenv("DOMAIN"),
		storage:    storage,
	}
}

func (s *Shortener) Shorten(userID, originalURL string) (string, error) {
	intlen, _ := strconv.Atoi(s.urlLen)
	b := make([]byte, intlen)
	for i := range b {
		b[i] = s.charset[s.randomizer.Intn(len(s.charset))]
	}
	shortURL := s.myDomain + string(b)
	err := s.storage.SaveURL(userID, shortURL, originalURL)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}
