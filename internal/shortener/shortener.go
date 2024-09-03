package shortener

import (
	"log"
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

// Creates new instance of Shortener service.
// Uses env variables to fill inner fields.
func NewShortener() *Shortener {
	return &Shortener{
		randomizer: rand.New(rand.NewSource(time.Now().UnixNano())),
		charset:    os.Getenv("CHARSET"),
		urlLen:     os.Getenv("LENGTH"),
		myDomain:   os.Getenv("DOMAIN"),
	}
}

// Creates random URL, which satisfies given conditions.
// Fatal if URL lenght condition is empty.
func (s *Shortener) Shorten() {
	intLen, err := strconv.Atoi(s.urlLen)
	if err != nil {
		log.Fatal(err)
	}
	b := make([]byte, intLen)
	for i := range b {
		b[i] = s.charset[s.randomizer.Intn(len(s.charset))]
	}
	s.ShortURL = s.myDomain + string(b)
}

// Generates new short URL.
// Used for preserving URL unique condition.
func (s *Shortener) Reshorten(short string) {
	s.Shorten()
	if s.ShortURL == short {
		s.Reshorten(s.ShortURL)
	}
}
