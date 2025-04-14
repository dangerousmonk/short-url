package models

import "time"

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type URLInfo struct {
	UUID        int       `json:"uuid"`
	OriginalURL string    `json:"original_url"`
	ShortURL    string    `json:"short_url"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

type APIBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type APIBatchModel struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
	ShortURL      string `json:"-"`
	Hash          string `json:"-"`
	UserID        string `json:"-"`
}

type APIGetUserURLsResponse struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}
