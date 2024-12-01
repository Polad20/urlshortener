package middleware

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

func brotliDecoder(data []byte) ([]byte, error) {
	newReader := brotli.NewReader(bytes.NewReader(data))
	decodedData, err := io.ReadAll(newReader)
	if err != nil {
		return nil, fmt.Errorf("Error reading the data")
	}
	return decodedData, nil
}

func brotliEncoder(data []byte) ([]byte, error) {
	var buffik bytes.Buffer
	newWriter := brotli.NewWriterLevel(&buffik, brotli.BestCompression)
	defer newWriter.Close()
	_, err := newWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("Error writing the data")
	}
	return buffik.Bytes(), nil
}

func middlewareBrotli(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
			next.ServeHTTP(w, r)
			return
		}
		readBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return

		}
		decoded, err := brotliDecoder(readBody)
		if err != nil {
			http.Error(w, "Error decoding Data", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(decoded))
		log.Printf("Request body was decoded with Brotli")
		next.ServeHTTP(w, r)
	})
}
