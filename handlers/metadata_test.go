package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"file-meta/config"
	"file-meta/internal/logger"
)

func TestMetadataHandler(t *testing.T) {
	cfg := &config.Config{
		Port:              "8080",
		MaxFileSizeMB:     20,
		RateLimitRequests: 10,
		RateLimitWindow:   time.Minute,
		LogLevel:          "info",
	}
	log := logger.New("info")

	tests := []struct {
		name           string
		fileContent    string
		fileName       string
		expectedStatus int
	}{
		{
			name:           "valid text file",
			fileContent:    "Hello, World!",
			fileName:       "test.txt",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty file",
			fileContent:    "",
			fileName:       "empty.txt",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("file", tt.fileName)
			if err != nil {
				t.Fatal(err)
			}
			io.WriteString(part, tt.fileContent)
			writer.Close()

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/v1/metadata", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler := MetadataHandler(cfg, log)
			handler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check content type for successful responses
			if tt.expectedStatus == http.StatusOK {
				if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
					t.Errorf("handler returned wrong content type: got %v want application/json", contentType)
				}
			}
		})
	}
}

func TestMetadataHandlerMissingFile(t *testing.T) {
	cfg := &config.Config{
		Port:              "8080",
		MaxFileSizeMB:     20,
		RateLimitRequests: 10,
		RateLimitWindow:   time.Minute,
		LogLevel:          "info",
	}
	log := logger.New("info")

	// Create empty multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/metadata", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	handler := MetadataHandler(cfg, log)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
