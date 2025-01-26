package model

type ShortenedURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
