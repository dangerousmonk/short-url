package memory

import (
	"context"

	"github.com/dangerousmonk/short-url/internal/models"
)

// GetURLData retrieves the original URL data by hash from memory
func (r *MemoryRepository) GetURLData(ctx context.Context, shortURL string) (urlData models.URLData, isExist bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	fullURL, isExist := r.MemoryStorage[shortURL]
	urlData = models.URLData{OriginalURL: fullURL, Active: true}
	return urlData, isExist
}
