package memory

import (
	"context"

	"github.com/dangerousmonk/short-url/internal/models"
)

// GetStats retrieves stats with stored URLs count
func (r *MemoryRepository) GetStats(ctx context.Context) (models.InternalStats, error) {
	stats := models.InternalStats{
		URLs:  len(r.MemoryStorage),
		Users: 0,
	}
	return stats, nil

}
