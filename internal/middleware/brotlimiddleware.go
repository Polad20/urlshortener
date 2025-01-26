package middleware

import (
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

type brotliResponseWriter struct {
	http.ResponseWriter
	Writer *brotli.Writer
}

func (brw *brotliResponseWriter) Write(b []byte) (int, error) {
	return brw.Writer.Write(b)
}

func (brw *brotliResponseWriter) Flush() error {
	return brw.Writer.Flush()
}

func MiddlewareBrotliEncoder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "br")
		w.Header().Del("Content-Length")
		bw := brotli.NewWriterLevel(w, brotli.BestCompression)
		defer bw.Close()
		brw := &brotliResponseWriter{ResponseWriter: w, Writer: bw}
		next.ServeHTTP(brw, r)
	})
}
