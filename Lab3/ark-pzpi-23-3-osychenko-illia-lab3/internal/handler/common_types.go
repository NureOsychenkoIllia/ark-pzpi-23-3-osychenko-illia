package handler

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error" example:"Error message"`
}

// MessageResponse represents a standard success message response
type MessageResponse struct {
	Message string `json:"message" example:"Operation successful"`
}
