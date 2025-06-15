package repository

import (
	"context"
	"database/sql"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/models"
)

//go:generate mockgen -package mocks -source types.go -destination ./mocks/mock_repository.go Repository
type Repository interface {
	// GetURLData retrieves the original URL data by hash
	GetURLData(ctx context.Context, shortURL string) (URLData models.URLData, isExist bool)
	// AddShortURL generates hash for provided URL and saves it along with original URL to internal storage
	AddShortURL(ctx context.Context, fullURL string, shortURL string, cfg *config.Config, userID string) (string, error)
	// Ping checks whether internal storage is up and running
	Ping(ctx context.Context) error
	// AddBatch generates hash for multiple URLS and saves it along with original URL to internal storage
	AddBatch(ctx context.Context, urls []models.APIBatchModel, cfg *config.Config, userID string) ([]models.APIBatchResponse, error)
	// GetUsersURLs retrieves all saved URLs by specific user
	GetUsersURLs(ctx context.Context, userID, baseURL string) ([]models.APIGetUserURLsResponse, error)
	// DeleteBatch marks multiple records as not active
	DeleteBatch(ctx context.Context, urls []string, userID string) error
}

type PostgresRepo struct {
	conn *sql.DB
}

func NewPostgresRepo(conn *sql.DB) *PostgresRepo {
	return &PostgresRepo{conn}
}
