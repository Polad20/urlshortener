package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	storage *Inmem = NewInmem()
)

func TestSaveData(t *testing.T) {
	tests := []struct {
		name  string
		og    string
		prep  int
		short string
		err   error
	}{
		{
			name:  "Positive SaveData",
			og:    "https://google.com",
			short: "search",
			err:   nil,
		},
		{
			name:  "Negative SaveData - og url exists",
			og:    "https://fangle.com",
			short: "push",
			prep:  1,
			err:   errNotUniqueOrigUrl,
		},
		{
			name:  "Negative SaveData - short url not unique",
			og:    "https://gigle.com",
			short: "break",
			prep:  2,
			err:   errNotUniqueShortUrl,
		},
	}
	for _, tc := range tests {
		switch tc.prep {
		case 1:
			storage.origUrls[tc.og] = 1
			storage.urlList[tc.short] = tc.og
		case 2:
			storage.urlList[tc.short] = tc.og
		}

		t.Run(tc.name, func(t *testing.T) {
			_, err := storage.SaveData(tc.og, tc.short)
			assert.Equal(t, tc.err, err)
			if tc.prep == 2 {
				return
			}
			assert.Contains(t, storage.origUrls, tc.og)
			assert.Contains(t, storage.urlList, tc.short)
		})
	}
}

func TestGetData(t *testing.T) {
	tests := []struct {
		name   string
		prep   int
		short  string
		result string
		err    error
	}{
		{
			name:   "Postitve GetData",
			prep:   1,
			result: "https://gigle.com",
			short:  "head",
		},
		{
			name:  "Negative GetData",
			prep:  2,
			short: "leg",
			err:   errNowSuchUrl,
		},
	}
	for _, tc := range tests {
		switch tc.prep {
		case 1:
			storage.SaveData("https://gigle.com", tc.short)
		}
		t.Run(tc.name, func(t *testing.T) {
			og, err := storage.GetData(tc.short)
			assert.Equal(t, tc.result, og)
			assert.Equal(t, tc.err, err)
			// v, ok := Storage.(storage)
		})
	}
}
