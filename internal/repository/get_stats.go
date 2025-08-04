package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dangerousmonk/short-url/internal/models"
)

// GetStats returns total number of stored URLs and users
func (r *PostgresRepo) GetStats(ctx context.Context) (models.InternalStats, error) {
	var stats models.InternalStats

	const getStatsQuery = `
	SELECT 
		COUNT(*) AS urls,
		COUNT(DISTINCT user_id) AS users
	FROM urls;`

	result := r.conn.QueryRowContext(ctx, getStatsQuery)
	err := result.Scan(&stats.URLs, &stats.Users)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.InternalStats{Users: 0, URLs: 0}, nil
		}
		return models.InternalStats{}, err
	}
	return stats, nil
}
