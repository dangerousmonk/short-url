package repository

import (
	"context"
	"fmt"
	"strings"
)

func (r *PostgresRepo) DeleteBatch(ctx context.Context, urls []string, userID string) error {
	var args []any

	placeholders := make([]string, len(urls))
	for i, url := range urls {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args = append(args, url)
	}
	args = append(args, userID)

	query := fmt.Sprintf(`
	 UPDATE urls
	 SET active=false
	 WHERE short_url IN (%s) AND user_id=$%d`,
		strings.Join(placeholders, ","),
		len(urls)+1,
	)

	_, err := r.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil

}
