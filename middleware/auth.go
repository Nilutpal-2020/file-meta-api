package middleware

import (
	"net/http"

	"file-meta/config"
	"file-meta/internal/logger"
)

// APIKeyAuth validates API key from request header
func APIKeyAuth(cfg *config.Config, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")

			if key == "" {
				log.Warn("Missing API key in request")
				http.Error(w, "Missing API key", http.StatusUnauthorized)
				return
			}

			if !cfg.APIKeys[key] {
				log.Warnf("Invalid API key attempted: %s", key[:min(len(key), 8)]+"...")
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
