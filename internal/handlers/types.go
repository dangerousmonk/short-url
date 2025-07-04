package handlers

import "github.com/dangerousmonk/short-url/internal/service"

// HTTPHandler represents base HTTP handler for the application
type HTTPHandler struct {
	service service.URLShortenerService
}

// NewHandler is a helper function to initialize new HTTP handler
func NewHandler(s service.URLShortenerService) *HTTPHandler {
	return &HTTPHandler{service: s}
}
