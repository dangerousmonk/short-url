package repository

import (
	"context"
	"time"
)

func (r *PostgresRepo) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := r.conn.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
