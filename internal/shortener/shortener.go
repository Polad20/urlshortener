package shortener

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/Polad20/urlshortener/config"
)

type Shortener struct {
	OriginalURL string `json:"originalurl"`
	ShortURL    string `json:"shorturl"`
	randomizer  *rand.Rand
	charset     string
	urlLen      string
	myDomain    string
}

func NewShortener() *Shortener {
	return &Shortener{
		randomizer: rand.New(rand.NewSource(time.Now().UnixNano())),
		charset:    config.AppConfig.URL.Charset,
		urlLen:     config.AppConfig.URL.Length,
		myDomain:   config.AppConfig.URL.Domain,
	}
}

func (s *Shortener) Shorten() string {
	intlen, _ := strconv.Atoi(s.urlLen)
	b := make([]byte, intlen)
	for i := range b {
		b[i] = s.charset[s.randomizer.Intn(len(s.charset))]
	}
	shortURL := s.myDomain + string(b)
	return shortURL
}
