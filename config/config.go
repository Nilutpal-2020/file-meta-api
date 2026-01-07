package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds application configuration
type Config struct {
	Port              string
	APIKeys           map[string]bool
	MaxFileSizeMB     int64
	RateLimitRequests int
	RateLimitWindow   time.Duration
	LogLevel          string
	Environment       string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Port:              getEnv("PORT", "8080"),
		MaxFileSizeMB:     getEnvAsInt("MAX_FILE_SIZE_MB", 20),
		RateLimitRequests: int(getEnvAsInt("RATE_LIMIT_REQUESTS", 10)),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		Environment:       getEnv("ENV", "development"),
	}

	// Parse rate limit window
	windowStr := getEnv("RATE_LIMIT_WINDOW", "1m")
	window, err := time.ParseDuration(windowStr)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_WINDOW: %w", err)
	}
	cfg.RateLimitWindow = window

	// Parse API keys
	apiKeysStr := os.Getenv("API_KEYS")
	if apiKeysStr == "" {
		return nil, fmt.Errorf("API_KEYS environment variable is required")
	}

	cfg.APIKeys = make(map[string]bool)
	for _, key := range strings.Split(apiKeysStr, ",") {
		key = strings.TrimSpace(key)
		if key != "" {
			cfg.APIKeys[key] = true
		}
	}

	if len(cfg.APIKeys) == 0 {
		return nil, fmt.Errorf("at least one API key is required")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if configuration values are valid
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("PORT cannot be empty")
	}

	if c.MaxFileSizeMB <= 0 {
		return fmt.Errorf("MAX_FILE_SIZE_MB must be positive")
	}

	if c.RateLimitRequests <= 0 {
		return fmt.Errorf("RATE_LIMIT_REQUESTS must be positive")
	}

	if c.RateLimitWindow <= 0 {
		return fmt.Errorf("RATE_LIMIT_WINDOW must be positive")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid LOG_LEVEL: must be one of debug, info, warn, error")
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as int64 or returns a default value
func getEnvAsInt(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}

	return value
}
