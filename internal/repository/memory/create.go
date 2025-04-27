package memory

import (
	"context"
	"strconv"
	"time"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/helpers"
	"github.com/dangerousmonk/short-url/internal/models"
)

func (r *MemoryRepository) AddShortURL(ctx context.Context, fullURL string, cfg *config.Config, userID string) (shortURL string, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for {
		shortURL, err = helpers.HashGenerator()
		if err != nil {
			return "", err
		}

		if _, exists := r.MemoryStorage[shortURL]; !exists {
			break
		}
	}

	r.MemoryStorage[shortURL] = fullURL
	urlData := models.URLData{UUID: strconv.Itoa(len(r.MemoryStorage)), ShortURL: shortURL, OriginalURL: fullURL, Active: true, CreatedAt: time.Now()}

	writer, err := NewWriter(cfg.StorageFilePath)
	if err != nil {
		return
	}
	defer writer.Close()

	if err = writer.WriteData(&urlData); err != nil {
		return
	}
	return shortURL, nil
}
