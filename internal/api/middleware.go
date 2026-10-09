package api

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// AuthConfig controls token verification in the API.
type AuthConfig struct {
	APIKey string
}

// RequireAuth enforces Bearer token authentication.
// If apiKey is empty, authentication is disabled (local dev mode on loopback).
// If apiKey is non-empty, every non-public request must provide a matching Bearer token.
func RequireAuth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no API key configured, bypass auth (local dev mode)
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization: Bearer <token> required")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format; expected Bearer <token>")
				return
			}

			providedToken := strings.TrimSpace(parts[1])
			if subtle.ConstantTimeCompare([]byte(providedToken), []byte(apiKey)) != 1 {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ContentTypeJSON sets the Content-Type header to application/json.
func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
