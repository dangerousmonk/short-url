package models

import "time"

// URLData represents common URL data from storage
type URLData struct {
	UUID        string    `json:"uuid"`
	OriginalURL string    `json:"original_url"`
	ShortURL    string    `json:"short_url"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

// APIBatchResponse represents required fields for APIShortenBatch HTTP handler
type APIBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// APIBatchModel represents common URL data
type APIBatchModel struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
	ShortURL      string `json:"-"`
	Hash          string `json:"-"`
	UserID        string `json:"-"`
}

// APIGetUserURLsResponse represents response structure for GetUserURLs HTTP handler
type APIGetUserURLsResponse struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	Hash        string `json:"-"`
}

// DeleteURLChannelMessage represents structure of delete message channel
type DeleteURLChannelMessage struct {
	URLs   []string
	UserID string
}
