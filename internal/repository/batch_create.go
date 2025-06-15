package repository

import (
	"context"
	"time"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/models"
)

// AddBatch generates hash for multiple URLS and saves it along with original URL to internal storage
func (r *PostgresRepo) AddBatch(ctx context.Context, urls []models.APIBatchModel, cfg *config.Config, userID string) ([]models.APIBatchResponse, error) {
	tx, err := r.conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res := make([]models.APIBatchResponse, 0, len(urls))

	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	for _, urlModel := range urls {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (original_url) DO NOTHING`,
			urlModel.Hash,
			urlModel.OriginalURL,
			userID,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, models.APIBatchResponse{CorrelationID: urlModel.CorrelationID, ShortURL: urlModel.ShortURL})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return res, nil
}
