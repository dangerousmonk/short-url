package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/helpers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgreSQLStorage struct {
	DB *sql.DB
}

type URLExistsError struct {
	ShortURL string
	Err      string
}

func (ps *PostgreSQLStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := ps.DB.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (ps *PostgreSQLStorage) GetFullURL(ctx context.Context, shortURL string) (fullURL string, isExist bool) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	row := ps.DB.QueryRowContext(ctx, `SELECT uuid, original_url, short_url, active, created_at FROM urls WHERE short_url=$1`, shortURL)
	var urlInfo models.URLInfo
	err := row.Scan(&urlInfo.UUID, &urlInfo.OriginalURL, &urlInfo.ShortURL, &urlInfo.Active, &urlInfo.CreatedAt)

	if err == nil {
		return urlInfo.OriginalURL, true
	}
	if err == sql.ErrNoRows {
		return "", false
	}
	logging.Log.Warnf("Error fetching URL | %v", err)
	return "", false
}

func (ps *PostgreSQLStorage) AddShortURL(ctx context.Context, fullURL string, cfg *config.Config, userID string) (shortURL string, err error) {
	shortURL, err = helpers.HashGenerator()
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	_, err = ps.DB.ExecContext(ctx, `INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)`, shortURL, fullURL, userID)
	if err == nil {
		return shortURL, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		err = ps.NewURLExistsError(fullURL, err)
	}
	return "", err
}

func (ps *PostgreSQLStorage) AddBatch(ctx context.Context, urls []models.APIBatchModel, cfg *config.Config, userID string) ([]models.APIBatchResponse, error) {
	tx, err := ps.DB.Begin()
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

func InitDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	logging.Log.Info("Database setup complete")
	return db, nil
}

func (error *URLExistsError) Error() string {
	return error.Err
}

func (ps *PostgreSQLStorage) NewURLExistsError(originalURL string, err error) *URLExistsError {
	var short string
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	row := ps.DB.QueryRowContext(ctx, `SELECT short_url FROM urls WHERE original_url = $1;`, originalURL)
	qErr := row.Scan(&short)
	if qErr != nil {
		return &URLExistsError{ShortURL: "", Err: "Error on quering existing url"}
	}
	return &URLExistsError{ShortURL: short, Err: "URL already exists"}
}

func (ps *PostgreSQLStorage) GetUsersURLs(ctx context.Context, userID, baseURL string) ([]models.APIGetUserURLsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	var resultRows []models.APIGetUserURLsResponse

	rows, err := ps.DB.QueryContext(ctx, `SELECT original_url, short_url FROM urls WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var urlResponse models.APIGetUserURLsResponse
		if err = rows.Scan(&urlResponse.OriginalURL, &urlResponse.ShortURL); err != nil {
			return nil, err
		}
		urlResponse.ShortURL = fmt.Sprintf("%s/%s", baseURL, urlResponse.ShortURL)
		resultRows = append(resultRows, urlResponse)

	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resultRows, nil
}
