package repository

import (
	"context"
	"time"
)

// Ping checks whether postgresql storage is up and running
func (r *PostgresRepo) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := r.conn.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
