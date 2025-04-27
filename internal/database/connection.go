package database

import (
	"context"
	"database/sql"

	"github.com/dangerousmonk/short-url/internal/logging"
)

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
