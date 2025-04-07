package shortener

import (
	"math/rand"
	"os"
	"strconv"
	"time"
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
		charset:    os.Getenv("CHARSET"),
		urlLen:     os.Getenv("LENGTH"),
		myDomain:   os.Getenv("DOMAIN"),
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
