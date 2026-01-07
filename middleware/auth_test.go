package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"file-meta/config"
	"file-meta/internal/logger"
)

func TestAPIKeyAuth(t *testing.T) {
	cfg := &config.Config{
		APIKeys: map[string]bool{
			"valid_key": true,
		},
	}
	log := logger.New("info")

	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "valid API key",
			apiKey:         "valid_key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid API key",
			apiKey:         "invalid_key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing API key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that returns 200 OK
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with auth middleware
			handler := APIKeyAuth(cfg, log)(nextHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
		})
	}
}
