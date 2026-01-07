package api

import (
	"encoding/json"
	"net/http"

	"file-meta/internal/models"
)

// Handler is the Vercel serverless function for /api/health
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	json.NewEncoder(w).Encode(models.HealthResponse{Status: "ok"})
}
