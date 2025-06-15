// Package service is used to describe URLShortenerService and to helper initialize service
package service

import (
	"context"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/repository"
)

// URLShortener is an interface for the business logic layer.
type URLShortener interface {
	// GetURLData retrieves the original URL data by hash
	GetURLData(ctx context.Context, shortURL string) (URLData models.URLData, isExist bool)
	// AddShortURL generates hash for provided URL and saves it along with original URL to internal storage
	AddShortURL(ctx context.Context, fullURL string, cfg *config.Config, userID string) (shortURL string, err error)
	// Ping checks whether internal storage is up and running
	Ping(ctx context.Context) error
	// AddBatch generates hash for multiple URLS and saves it along with original URL to internal storage
	AddBatch(ctx context.Context, urls []models.APIBatchModel, cfg *config.Config, userID string) ([]models.APIBatchResponse, error)
	// GetUsersURLs retrieves all saved URLs by specific user
	GetUsersURLs(ctx context.Context, userID, baseURL string) ([]models.APIGetUserURLsResponse, error)
	// DeleteBatch marks multiple records as not active
	DeleteBatch(ctx context.Context, urls []string, userID string) error
	// FlushDeleteMessages sends messages for deletion
	FlushDeleteMessages()
}

// URLShortenerService is a struct that describes service.
type URLShortenerService struct {
	Repo  repository.Repository
	DelCh chan models.DeleteURLChannelMessage
	Cfg   *config.Config
}

// NewShortenerService is a function used to initialize new URLShortenerService
func NewShortenerService(r repository.Repository, cfg *config.Config, delCh chan models.DeleteURLChannelMessage) *URLShortenerService {
	service := URLShortenerService{Repo: r, Cfg: cfg, DelCh: delCh}
	return &service
}
