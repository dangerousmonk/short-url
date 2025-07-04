package service

import (
	"context"
	"time"
)

// Ping calls repository to check if storage is alive.
func (s *URLShortenerService) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	return s.Repo.Ping(ctx)
}
