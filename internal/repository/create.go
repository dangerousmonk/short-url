package repository

import (
	"context"
	"errors"
	"time"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type URLExistsError struct {
	ShortURL string
	URL      string
	Err      string
}

func (err *URLExistsError) Error() string {
	return err.Err
}

func (r *PostgresRepo) NewURLExistsError(originalURL string, err error) *URLExistsError {
	var short string
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	row := r.conn.QueryRowContext(ctx, `SELECT short_url FROM urls WHERE original_url = $1;`, originalURL)
	qErr := row.Scan(&short)
	if qErr != nil {
		return &URLExistsError{URL: originalURL, ShortURL: "", Err: "Error on querying existing url"}
	}
	return &URLExistsError{URL: originalURL, ShortURL: short, Err: "URL already exists"}
}

func (r *PostgresRepo) AddShortURL(ctx context.Context, fullURL string, shortURL string, cfg *config.Config, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	_, err := r.conn.ExecContext(ctx, `INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)`, shortURL, fullURL, userID)
	if err == nil {
		return shortURL, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		err = r.NewURLExistsError(fullURL, err)
	}
	return "", err
}
