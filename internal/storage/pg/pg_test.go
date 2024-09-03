package storage

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var storage *Postgres = NewPostgres()

func generateRandomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	randomizer := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = charset[randomizer.Intn(len(charset))]
	}
	return string(b)
}

func TestSaveData(t *testing.T) {

	tests := []struct {
		name  string
		og    string
		short string
		err   error
	}{
		{
			name:  "Negative SaveData not unique og",
			og:    "https://google.com",
			short: "search",
			err:   errNotUnique,
		},
		{
			name:  "Positive SaveData",
			og:    generateRandomString(10),
			short: generateRandomString(5),
		},
		{
			name:  "Negative SaveData not unique short",
			og:    generateRandomString(7),
			short: "search",
			err:   errNotUniqueShortUrl,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := storage.SaveData(tc.og, tc.short)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				assert.Equal(t, tc.short, res)
			}
		})
	}
}

func TestGetData(t *testing.T) {
	tests := []struct {
		name   string
		short  string
		result string
		err    error
	}{
		{
			name:   "Positive SaveData",
			short:  "search",
			result: "https://google.com",
		},
		{
			name:  "Negative GetData",
			short: "lkjasdf",
			err:   errNowSuchUrl,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			og, err := storage.GetData(tc.short)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.result, og)
		})
	}
}
