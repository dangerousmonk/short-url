package memory

import (
	"context"
	"errors"

	"github.com/dangerousmonk/short-url/internal/models"
)

// GetUsersURLs is not supported by MemoryRepository
func (r *MemoryRepository) GetUsersURLs(ctx context.Context, userID, baseURL string) ([]models.APIGetUserURLsResponse, error) {
	return nil, errors.New("mapStorage doesnt support GetUsersURLs method")
}
