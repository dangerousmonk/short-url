package service

import (
	"context"
	"errors"

	"github.com/dangerousmonk/short-url/internal/helpers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/repository"
)

var (
	ErrURLInvalid = errors.New("url: provided url is invalid")
	ErrURLExists  = errors.New("url: url already exists")
	ErrSaveFailed = errors.New("url: failed to save URL")
)

func (s *URLShortenerService) AddURL(url string, ctx context.Context, userID string) (string, error) {
	if !helpers.IsURLValid(url) {
		return "", ErrURLInvalid
	}

	shortURL, err := s.Repo.AddShortURL(ctx, url, s.Cfg, userID)
	if err != nil {
		logging.Log.Warnf("s:AddURL error on inserting URL | %v", err)
		var existsErr *repository.URLExistsError

		if errors.As(err, &existsErr) {
			logging.Log.Warnf("s:AddURL URLExistsError | %v", existsErr.URL)
			return s.Cfg.BaseURL + "/" + existsErr.ShortURL, ErrURLExists
		}
		return "", ErrSaveFailed
	}
	return s.Cfg.BaseURL + "/" + shortURL, nil
}
