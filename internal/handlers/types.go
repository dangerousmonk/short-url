package handlers

import "github.com/dangerousmonk/short-url/internal/service"

type HTTPHandler struct {
	service service.URLShortenerService
}

func NewHandler(s service.URLShortenerService) *HTTPHandler {
	return &HTTPHandler{service: s}
}
