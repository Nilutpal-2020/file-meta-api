package handlers

import (
	"encoding/json"
	"net/http"

	"file-meta/config"
	"file-meta/internal/logger"
	"file-meta/internal/metadata"
	"file-meta/middleware"
)

// MetadataHandler handles file metadata extraction requests
func MetadataHandler(cfg *config.Config, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetRequestID(r.Context())

		// Check Content-Length before parsing
		maxBytes := cfg.MaxFileSizeMB << 20 // Convert MB to bytes
		if r.ContentLength > maxBytes {
			log.Warnf("[%s] File too large: %d bytes", requestID, r.ContentLength)
			http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
			return
		}

		err := r.ParseMultipartForm(maxBytes)
		if err != nil {
			log.Errorf("[%s] Failed to parse multipart form: %v", requestID, err)
			http.Error(w, "File too large or invalid", http.StatusRequestEntityTooLarge)
			return
		}
		defer r.MultipartForm.RemoveAll()

		file, header, err := r.FormFile("file")
		if err != nil {
			log.Warnf("[%s] Invalid file in request: %v", requestID, err)
			http.Error(w, "Invalid file parameter", http.StatusBadRequest)
			return
		}
		defer file.Close()

		log.Debugf("[%s] Processing file: %s (%d bytes)", requestID, header.Filename, header.Size)

		result, err := metadata.Extract(file, header)
		if err != nil {
			log.Errorf("[%s] Failed to extract metadata: %v", requestID, err)
			http.Error(w, "Failed to extract metadata", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Errorf("[%s] Failed to encode response: %v", requestID, err)
		}

		log.Infof("[%s] Successfully processed file: %s", requestID, header.Filename)
	}
}
