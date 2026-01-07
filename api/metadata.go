package api

import (
	"context"
	"net/http"
	"sync"

	"file-meta/config"
	"file-meta/handlers"
	"file-meta/internal/logger"
	"file-meta/middleware"

	"github.com/redis/go-redis/v9"
)

var (
	cfg         *config.Config
	log         *logger.Logger
	redisClient *redis.Client
	once        sync.Once
	initErr     error
)

// Initialize config, logger, and Redis client (only once)
func initialize() {
	once.Do(func() {
		// Load configuration
		cfg, initErr = config.Load()
		if initErr != nil {
			return
		}

		// Initialize logger
		log = logger.New(cfg.LogLevel)

		// Initialize Redis client
		if cfg.RedisURL != "" {
			// Use Redis URL if provided
			opt, err := redis.ParseURL(cfg.RedisURL)
			if err != nil {
				initErr = err
				return
			}
			redisClient = redis.NewClient(opt)
		} else {
			// Use individual Redis settings
			redisClient = redis.NewClient(&redis.Options{
				Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
				Password: cfg.RedisPassword,
				DB:       cfg.RedisDB,
			})
		}

		// Test Redis connection
		ctx := context.Background()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Warnf("Redis connection failed, rate limiting will be in-memory: %v", err)
		} else {
			log.Info("Successfully connected to Redis")
		}
	})
}

// Handler is the Vercel serverless function for /api/metadata
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize once
	initialize()
	if initErr != nil {
		log.Errorf("Initialization error: %v", initErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// CORS preflight
	if r.Method == http.MethodOptions {
		middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})).ServeHTTP(w, r)
		return
	}

	// Create handler with middleware chain
	handler := middleware.Recovery(log)(
		middleware.RequestLogger(log)(
			middleware.RedisRateLimit(cfg, log, redisClient)(
				middleware.APIKeyAuth(cfg, log)(
					http.HandlerFunc(handlers.MetadataHandler(cfg, log)),
				),
			),
		),
	)

	// Wrap with CORS
	middleware.CORS(handler).ServeHTTP(w, r)
}
