package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Set required environment variables
	os.Setenv("API_KEYS", "test_key_1,test_key_2")
	os.Setenv("PORT", "9090")
	os.Setenv("MAX_FILE_SIZE_MB", "50")
	os.Setenv("RATE_LIMIT_REQUESTS", "20")
	os.Setenv("RATE_LIMIT_WINDOW", "2m")
	os.Setenv("LOG_LEVEL", "debug")

	defer func() {
		os.Unsetenv("API_KEYS")
		os.Unsetenv("PORT")
		os.Unsetenv("MAX_FILE_SIZE_MB")
		os.Unsetenv("RATE_LIMIT_REQUESTS")
		os.Unsetenv("RATE_LIMIT_WINDOW")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %v, want 9090", cfg.Port)
	}

	if cfg.MaxFileSizeMB != 50 {
		t.Errorf("MaxFileSizeMB = %v, want 50", cfg.MaxFileSizeMB)
	}

	if cfg.RateLimitRequests != 20 {
		t.Errorf("RateLimitRequests = %v, want 20", cfg.RateLimitRequests)
	}

	if cfg.RateLimitWindow != 2*time.Minute {
		t.Errorf("RateLimitWindow = %v, want 2m", cfg.RateLimitWindow)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}

	if !cfg.APIKeys["test_key_1"] || !cfg.APIKeys["test_key_2"] {
		t.Errorf("APIKeys not loaded correctly: %v", cfg.APIKeys)
	}
}

func TestLoadMissingAPIKeys(t *testing.T) {
	os.Unsetenv("API_KEYS")

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error when API_KEYS is missing")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Port:              "8080",
				MaxFileSizeMB:     20,
				RateLimitRequests: 10,
				RateLimitWindow:   time.Minute,
				LogLevel:          "info",
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: &Config{
				Port:              "8080",
				MaxFileSizeMB:     20,
				RateLimitRequests: 10,
				RateLimitWindow:   time.Minute,
				LogLevel:          "invalid",
			},
			wantErr: true,
		},
		{
			name: "negative file size",
			config: &Config{
				Port:              "8080",
				MaxFileSizeMB:     -1,
				RateLimitRequests: 10,
				RateLimitWindow:   time.Minute,
				LogLevel:          "info",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
