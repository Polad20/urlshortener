package shortener

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var s *Shortener = NewShortener()

func TestShorten(t *testing.T) {
	ogURL := "https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go"
	s.OriginalURL = ogURL
	s.Shorten()
	assert.NotEqual(t, "", s.ShortURL)
}
