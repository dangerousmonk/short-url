package service

import (
	"context"

	"github.com/dangerousmonk/short-url/internal/models"
)

// GetUsersURLs returns all URLs saved by user.
func (s *URLShortenerService) GetUsersURLs(ctx context.Context, userID string) ([]models.APIGetUserURLsResponse, error) {
	userURLs, err := s.Repo.GetUsersURLs(ctx, userID, s.Cfg.BaseURL)
	return userURLs, err
}
