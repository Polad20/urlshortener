package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	storage "github.com/Polad20/urlshortener/internal/storage/inmem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, jsonBody []byte) (int, string) {

	req, err := http.NewRequest(method, ts.URL+path, bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func TestRouter(t *testing.T) {
	repo := storage.NewInmem()
	if _, err := repo.SaveData("https://gigle.com", "short"); err != nil {
		panic(err)
	}
	r := NewHandler(repo)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		status int
		resp   string
	}

	tests := []struct {
		name     string
		method   string
		path     string
		jsonBody []byte
		want     want
	}{
		{
			name:     "Positive SaveURL",
			method:   http.MethodPost,
			path:     "/",
			jsonBody: []byte(`{"originalurl":"https://google.com"}`),
			want: want{
				status: 200,
			},
		},
		{
			name:     "Negative SaveURL",
			method:   http.MethodPost,
			path:     "/",
			jsonBody: []byte(``),
			want: want{
				status: 400,
			},
		},
		{
			name:     "Positive GetURL",
			method:   http.MethodGet,
			path:     "/",
			jsonBody: []byte(`{"shorturl":"short"}`),
			want: want{
				status: 200,
				resp:   "\"https://gigle.com\"\n",
			},
		},
		{
			name:     "Negative GetURL - no such url",
			method:   http.MethodGet,
			path:     "/",
			jsonBody: []byte(`{"shorturl":"random"}`),
			want: want{
				status: 404,
			},
		},
		{
			name:     "Negative GetURL - bad req",
			method:   http.MethodGet,
			path:     "/",
			jsonBody: []byte(``),
			want: want{
				status: 400,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, respBody := testRequest(t, ts, tc.method, tc.path, tc.jsonBody)
			assert.Equal(t, tc.want.status, status)
			if tc.want.resp != "" {
				assert.Equal(t, tc.want.resp, respBody)

			}
		})
	}
}
