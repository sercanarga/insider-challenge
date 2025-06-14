package handler

// StatusResponse represents a simple status response
type StatusResponse struct {
	Status string `json:"status"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
