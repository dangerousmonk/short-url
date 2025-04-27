package service

import (
	"context"
	"time"
)

func (s *URLShortenerService) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	return s.Repo.Ping(ctx)
}
