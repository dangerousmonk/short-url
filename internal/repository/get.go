package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
)

func (r *PostgresRepo) GetURLData(ctx context.Context, shortURL string) (URLData models.URLData, isExist bool) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	const selectFields = "original_url, short_url, active"
	row := r.conn.QueryRowContext(ctx, `SELECT `+selectFields+` FROM urls WHERE short_url=$1`, shortURL)

	var urlData models.URLData
	err := row.Scan(&urlData.OriginalURL, &urlData.ShortURL, &urlData.Active)

	if err == nil {
		return urlData, true
	}
	if err == sql.ErrNoRows {
		return urlData, false
	}
	logging.Log.Warnf("Error fetching URL | %v", err)
	return urlData, false
}
