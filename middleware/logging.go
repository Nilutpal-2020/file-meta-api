package middleware

import (
	"context"
	"net/http"
	"time"

	"file-meta/internal/logger"

	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "requestID"

// RequestLogger logs HTTP requests with request ID and timing
func RequestLogger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID
			requestID := uuid.New().String()

			// Add request ID to context
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)

			// Add request ID to response headers
			w.Header().Set("X-Request-ID", requestID)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Log request
			log.Infof("[%s] %s %s - Started", requestID, r.Method, r.URL.Path)

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log response
			duration := time.Since(start)
			log.Infof("[%s] %s %s - Completed %d in %v",
				requestID, r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
