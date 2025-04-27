package service

import (
	"context"

	"github.com/dangerousmonk/short-url/internal/models"
)

func (s *URLShortenerService) GetURLData(ctx context.Context, hash string) (models.URLData, bool) {
	urlData, isExist := s.Repo.GetURLData(ctx, hash)
	return urlData, isExist
}
