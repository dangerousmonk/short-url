package service

import (
	"context"

	"github.com/dangerousmonk/short-url/internal/models"
)

// GetURLData is used to retreive info about single URL by provided hash.
func (s *URLShortenerService) GetURLData(ctx context.Context, hash string) (models.URLData, bool) {
	urlData, isExist := s.Repo.GetURLData(ctx, hash)
	return urlData, isExist
}
