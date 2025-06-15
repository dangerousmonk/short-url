package service

import (
	"context"
	"errors"

	"github.com/dangerousmonk/short-url/internal/helpers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
)

// Errors that might occur during BatchCreate operation.
var (
	ErrTooManyURLs     = errors.New("urls: too many urls provided")
	ErrNoValidURLs     = errors.New("urls: at least one valid URL required")
	ErrHashFailed      = errors.New("hash: failed to generate hash")
	ErrSaveBatchFailed = errors.New("repo: failed to save data")
)

// BatchCreate is used to create multiple short URLs.
func (s *URLShortenerService) BatchCreate(urls []models.APIBatchModel, ctx context.Context, userID string) ([]models.APIBatchResponse, error) {
	var validURLs []models.APIBatchModel

	if len(urls) > s.Cfg.MaxURLsBatchSize {
		logging.Log.Warnf("s:BatchCreate too many URLs=%v, allowed size=%v", len(urls), s.Cfg.MaxURLsBatchSize)
		return nil, ErrTooManyURLs
	}

	for _, url := range urls {
		if helpers.IsURLValid(url.OriginalURL) {
			validURLs = append(validURLs, url)
		}
	}

	if len(validURLs) == 0 {
		logging.Log.Warnf("s:BatchCreate no valid URLs=%v", urls)
		return nil, ErrNoValidURLs
	}

	for idx := range validURLs {
		hash, err := helpers.HashGenerator()
		if err != nil {
			logging.Log.Warnf("s:BatchCreate failed generate hash err=%v", err)
			return nil, ErrHashFailed
		}
		short := s.Cfg.BaseURL + "/" + hash
		validURLs[idx].ShortURL = short
		validURLs[idx].Hash = hash
	}

	resp, err := s.Repo.AddBatch(ctx, validURLs, s.Cfg, userID)
	if err != nil {
		logging.Log.Warnf("s:BatchCreate failed to save URLs err=%v", err)
		return nil, ErrSaveBatchFailed
	}
	return resp, nil

}
