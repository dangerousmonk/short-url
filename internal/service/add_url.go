package service

import (
	"context"
	"errors"

	"github.com/dangerousmonk/short-url/internal/helpers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/repository"
)

// Errors that might occur during AddURL operation.
var (
	ErrURLInvalid = errors.New("url: provided url is invalid")
	ErrURLExists  = errors.New("url: url already exists")
	ErrSaveFailed = errors.New("url: failed to save URL")
)

// AddURL is used to create single short URL.
func (s *URLShortenerService) AddURL(ctx context.Context, url string, userID string) (string, error) {
	if !helpers.IsURLValid(url) {
		return "", ErrURLInvalid
	}

	shortURL, err := helpers.HashGenerator()
	if err != nil {
		return "", ErrSaveFailed
	}

	shortURL, err = s.Repo.AddShortURL(ctx, url, shortURL, s.Cfg, userID)
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
