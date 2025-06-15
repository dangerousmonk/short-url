package memory

import (
	"context"
	"strconv"
	"time"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/models"
)

func (r *MemoryRepository) AddShortURL(ctx context.Context, fullURL string, shortURL string, cfg *config.Config, userID string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.MemoryStorage[shortURL] = fullURL
	urlData := models.URLData{UUID: strconv.Itoa(len(r.MemoryStorage)), ShortURL: shortURL, OriginalURL: fullURL, Active: true, CreatedAt: time.Now()}

	writer, err := NewWriter(cfg.StorageFilePath)
	if err != nil {
		return "", err
	}
	defer writer.Close()

	if err = writer.WriteData(&urlData); err != nil {
		return "", err
	}
	return shortURL, nil
}
