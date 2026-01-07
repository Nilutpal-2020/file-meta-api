package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"file-meta/config"
	"file-meta/internal/logger"
)

type client struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

var (
	mu             sync.RWMutex
	clients        = make(map[string]*client)
	cleanupStarted sync.Once
)

// RateLimit implements token bucket rate limiting
func RateLimit(cfg *config.Config, log *logger.Logger) func(http.Handler) http.Handler {
	// Start cleanup goroutine only once
	cleanupStarted.Do(func() {
		go cleanupExpiredClients(cfg.RateLimitWindow, log)
	})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			now := time.Now()

			mu.Lock()
			c, exists := clients[key]
			if !exists {
				c = &client{
					tokens:     cfg.RateLimitRequests,
					lastRefill: now,
				}
				clients[key] = c
			}
			mu.Unlock()

			c.mu.Lock()
			defer c.mu.Unlock()

			// Refill tokens if window has passed
			if now.Sub(c.lastRefill) > cfg.RateLimitWindow {
				c.tokens = cfg.RateLimitRequests
				c.lastRefill = now
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RateLimitRequests))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", c.tokens))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", c.lastRefill.Add(cfg.RateLimitWindow).Unix()))

			if c.tokens <= 0 {
				log.Warnf("Rate limit exceeded for API key: %s", key[:min(len(key), 8)]+"...")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			c.tokens--
			next.ServeHTTP(w, r)
		})
	}
}

// cleanupExpiredClients removes expired clients from memory
func cleanupExpiredClients(window time.Duration, log *logger.Logger) {
	ticker := time.NewTicker(window * 2)
	defer ticker.Stop()

	for range ticker.C {
		mu.Lock()
		now := time.Now()
		for key, c := range clients {
			c.mu.Lock()
			if now.Sub(c.lastRefill) > window*3 {
				delete(clients, key)
				log.Debugf("Cleaned up expired client: %s", key[:min(len(key), 8)]+"...")
			}
			c.mu.Unlock()
		}
		mu.Unlock()
	}
}
