package service

import (
	"context"

	"github.com/dangerousmonk/short-url/internal/models"
)

// GetStats is used to retreive stats with stored URLs count and unique user_id count
func (s *URLShortenerService) GetStats(ctx context.Context) (models.InternalStats, error) {
	stats, err := s.Repo.GetStats(ctx)
	if err != nil {
		return models.InternalStats{}, err
	}
	return stats, nil
}
