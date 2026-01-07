package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"file-meta/config"
	"file-meta/internal/logger"
)

func TestRateLimit(t *testing.T) {
	cfg := &config.Config{
		RateLimitRequests: 2,
		RateLimitWindow:   time.Second,
	}
	log := logger.New("info")

	// Create a test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with rate limit middleware
	handler := RateLimit(cfg, log)(nextHandler)

	// Test within limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-API-Key", "test_key")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("request %d: got status %v, want %v", i+1, status, http.StatusOK)
		}
	}

	// Test exceeding limit
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "test_key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusTooManyRequests {
		t.Errorf("rate limited request: got status %v, want %v", status, http.StatusTooManyRequests)
	}

	// Check rate limit headers
	if rr.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("X-RateLimit-Limit header not set")
	}
}

func TestRateLimitRefill(t *testing.T) {
	// Clear any existing clients from previous tests
	mu.Lock()
	clients = make(map[string]*client)
	mu.Unlock()

	cfg := &config.Config{
		RateLimitRequests: 1,
		RateLimitWindow:   100 * time.Millisecond,
	}
	log := logger.New("info")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RateLimit(cfg, log)(nextHandler)

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("X-API-Key", "test_refill_key")
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if status := rr1.Code; status != http.StatusOK {
		t.Errorf("first request: got status %v, want %v", status, http.StatusOK)
	}

	// Second request should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-API-Key", "test_refill_key")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if status := rr2.Code; status != http.StatusTooManyRequests {
		t.Errorf("second request: got status %v, want %v", status, http.StatusTooManyRequests)
	}

	// Wait for token refill
	time.Sleep(150 * time.Millisecond)

	// Third request should succeed after refill
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.Header.Set("X-API-Key", "test_refill_key")
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)

	if status := rr3.Code; status != http.StatusOK {
		t.Errorf("third request after refill: got status %v, want %v", status, http.StatusOK)
	}
}
