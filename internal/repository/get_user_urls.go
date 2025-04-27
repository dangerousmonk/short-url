package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/dangerousmonk/short-url/internal/models"
)

func (r *PostgresRepo) GetUsersURLs(ctx context.Context, userID, baseURL string) ([]models.APIGetUserURLsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	var resultRows []models.APIGetUserURLsResponse

	rows, err := r.conn.QueryContext(ctx, `SELECT original_url, short_url FROM urls WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var urlResponse models.APIGetUserURLsResponse
		if err = rows.Scan(&urlResponse.OriginalURL, &urlResponse.Hash); err != nil {
			return nil, err
		}
		urlResponse.ShortURL = fmt.Sprintf("%s/%s", baseURL, urlResponse.Hash)
		resultRows = append(resultRows, urlResponse)

	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resultRows, nil
}
