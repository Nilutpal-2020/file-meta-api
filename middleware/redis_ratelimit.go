package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"file-meta/config"
	"file-meta/internal/logger"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimit implements distributed rate limiting using Redis
func RedisRateLimit(cfg *config.Config, log *logger.Logger, redisClient *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()
			key := r.Header.Get("X-API-Key")
			now := time.Now()

			// Redis key for this API key
			rateLimitKey := fmt.Sprintf("ratelimit:%s", key)

			// Try to get current token count
			tokens, err := redisClient.Get(ctx, rateLimitKey).Int()
			if err == redis.Nil {
				// Key doesn't exist, initialize with max tokens
				tokens = cfg.RateLimitRequests
				err = redisClient.Set(ctx, rateLimitKey, tokens, cfg.RateLimitWindow).Err()
				if err != nil {
					log.Errorf("Redis error: %v", err)
					// Fallback: allow request if Redis is down
					next.ServeHTTP(w, r)
					return
				}
			} else if err != nil {
				log.Errorf("Redis error: %v", err)
				// Fallback: allow request if Redis is down
				next.ServeHTTP(w, r)
				return
			}

			// Get TTL to calculate reset time
			ttl, err := redisClient.TTL(ctx, rateLimitKey).Result()
			if err != nil {
				log.Errorf("Redis TTL error: %v", err)
				ttl = cfg.RateLimitWindow
			}

			resetTime := now.Add(ttl).Unix()

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RateLimitRequests))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, tokens-1)))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))

			// Check if rate limited
			if tokens <= 0 {
				log.Warnf("Rate limit exceeded for API key: %s", key[:min(len(key), 8)]+"...")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// Decrement tokens
			newTokens, err := redisClient.Decr(ctx, rateLimitKey).Result()
			if err != nil {
				log.Errorf("Redis decrement error: %v", err)
				// Fallback: allow request if Redis is down
				next.ServeHTTP(w, r)
				return
			}

			// If this was the first decrement after initialization, set expiry
			if newTokens == int64(cfg.RateLimitRequests-1) {
				redisClient.Expire(ctx, rateLimitKey, cfg.RateLimitWindow)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
