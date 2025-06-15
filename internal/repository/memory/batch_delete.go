package memory

import (
	"context"
	"errors"
)

// DeleteBatch is not supported by MemoryRepository
func (r *MemoryRepository) DeleteBatch(ctx context.Context, urls []string, userID string) error {
	return errors.New("mapStorage doesnt support DeleteBatch method")
}
