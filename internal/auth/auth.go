package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Auth struct {
	key []byte
}

func New(key []byte) *Auth {
	return &Auth{
		key: key,
	}
}

func (a *Auth) signCookie(userID []byte) (string, error) {
	h := hmac.New(sha256.New, a.key)
	h.Write(userID)
	signature := h.Sum(nil)
	return fmt.Sprintf("%s.%s", hex.EncodeToString(userID), base64.RawURLEncoding.EncodeToString(signature)), nil
}

func (a *Auth) checkSignature(userID []byte, signature []byte) bool {
	h := hmac.New(sha256.New, a.key)
	h.Write(userID)
	calculatedSignature := h.Sum(nil)
	return hmac.Equal(signature, calculatedSignature)
}

func (a *Auth) setCookie(w http.ResponseWriter, name string, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
	}
	http.SetCookie(w, cookie)
}

func generateRandom(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (a *Auth) MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("userID")
		if err == http.ErrNoCookie {
			userID, err := generateRandom(32)
			if err != nil {
				log.Printf("Error generating userID: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			signedCookie, err := a.signCookie(userID)
			if err != nil {
				log.Printf("Error signing cookie: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			a.setCookie(w, "userID", signedCookie)
			ctx := context.WithValue(r.Context(), "userID", hex.EncodeToString(userID))
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		parts := strings.Split(cookie.Value, ".")
		if len(parts) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Bad cookie format: %v", cookie.Value)
			return
		}
		userID := parts[0]
		originalUserIDBytes, err := hex.DecodeString(userID)
		if err != nil {
			log.Printf("Error decoding userID hex from cookie: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		signature, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			log.Printf("Error decoding signature: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ok := a.checkSignature(originalUserIDBytes, signature)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			log.Printf("Signature not valid")
			return
		}
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
