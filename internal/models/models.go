package models

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status string `json:"status"`
}
