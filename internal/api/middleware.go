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

// RequireAuth enforces Bearer token authentication or trusted session cookie authentication.
// If apiKey is empty, authentication is disabled (local dev mode on loopback).
// If apiKey is non-empty, every non-public request must provide either a matching
// Bearer token or a valid session cookie from an authenticated BFF/reverse-proxy boundary.
func RequireAuth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no API key configured, bypass auth (local dev mode)
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check Bearer authorization header (CLI, API client, authenticated reverse proxy)
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					providedToken := strings.TrimSpace(parts[1])
					if subtle.ConstantTimeCompare([]byte(providedToken), []byte(apiKey)) == 1 {
						next.ServeHTTP(w, r)
						return
					}
					writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid API key")
					return
				}
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format; expected Bearer <token>")
				return
			}

			// Check session cookie (for same-origin browser sessions mediated by a BFF/proxy)
			if cookie, err := r.Cookie("stellaryard_session"); err == nil && cookie.Value != "" {
				if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(apiKey)) == 1 {
					next.ServeHTTP(w, r)
					return
				}
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid session token")
				return
			}

			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization: Bearer <token> or valid session required")
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
