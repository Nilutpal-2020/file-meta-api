package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"file-meta/config"
	"file-meta/handlers"
	"file-meta/internal/logger"
	"file-meta/internal/models"
	"file-meta/middleware"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel)
	log.Infof("Starting file-meta server in %s mode", cfg.Environment)

	// Initialize Redis client (optional)
	var redisClient *redis.Client
	if cfg.RedisURL != "" || cfg.RedisHost != "" {
		if cfg.RedisURL != "" {
			// Use Redis URL if provided
			opt, err := redis.ParseURL(cfg.RedisURL)
			if err != nil {
				log.Warnf("Failed to parse Redis URL: %v", err)
			} else {
				redisClient = redis.NewClient(opt)
			}
		} else {
			// Use individual Redis settings
			redisClient = redis.NewClient(&redis.Options{
				Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
				Password: cfg.RedisPassword,
				DB:       cfg.RedisDB,
			})
		}

		// Test Redis connection
		if redisClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := redisClient.Ping(ctx).Err(); err != nil {
				log.Warnf("Redis connection failed, using in-memory rate limiting: %v", err)
				redisClient = nil
			} else {
				log.Info("Successfully connected to Redis for distributed rate limiting")
			}
		}
	}

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.HealthResponse{Status: "ok"})
	})

	// Choose rate limiting strategy
	var rateLimitMiddleware func(http.Handler) http.Handler
	if redisClient != nil {
		rateLimitMiddleware = middleware.RedisRateLimit(cfg, log, redisClient)
	} else {
		rateLimitMiddleware = middleware.RateLimit(cfg, log)
	}

	// Metadata endpoint with middleware chain
	handler := middleware.Recovery(log)(
		middleware.RequestLogger(log)(
			rateLimitMiddleware(
				middleware.APIKeyAuth(cfg, log)(
					http.HandlerFunc(handlers.MetadataHandler(cfg, log)),
				),
			),
		),
	)

	mux.Handle("/v1/metadata", middleware.CORS(handler))

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Infof("Server listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Server shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server stopped gracefully")
}
