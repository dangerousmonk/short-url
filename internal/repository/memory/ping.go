package memory

import "context"

// Ping checks whether internal storage is up and running
func (r *MemoryRepository) Ping(ctx context.Context) error {
	return nil
}
