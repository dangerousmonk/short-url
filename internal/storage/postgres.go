package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/helpers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
)

type PostgreSQLStorage struct {
	DB *sql.DB
}

func (ps *PostgreSQLStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	if err := ps.DB.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (ps *PostgreSQLStorage) GetFullURL(shortURL string) (fullURL string, isExist bool) {
	row := ps.DB.QueryRow(`SELECT uuid, original_url, short_url, active, created_at FROM urls WHERE short_url=$1`, shortURL)
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

func (ps *PostgreSQLStorage) AddShortURL(fullURL string, cfg *config.Config) (shortURL string, err error) {
	shortURL, err = helpers.HashGenerator()
	if err != nil {
		return
	}

	_, err = ps.DB.Exec(`INSERT INTO urls (short_url, original_url) VALUES ($1, $2)`, shortURL, fullURL)
	if err != nil {
		return
	}
	return shortURL, nil
}

func (ps *PostgreSQLStorage) AddBatch(urls []models.APIBatchModel, cfg *config.Config) error {
	tx, err := ps.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, urlModel := range urls {
		_, err = tx.Exec(`INSERT INTO urls (short_url, original_url) VALUES ($1, $2)`, urlModel.Hash, urlModel.OriginalURL)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func InitDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	err = initSchema(ctx, db)
	if err != nil {
		return nil, err
	}
	logging.Log.Info("Database setup complete")
	return db, nil
}

func initSchema(ctx context.Context, db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS urls (
        uuid  BIGSERIAL primary key,
        original_url TEXT NOT NULL,
        short_url VARCHAR(50) NOT NULL,
		active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return nil
}
