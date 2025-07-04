package memory

import (
	"context"
	"strconv"
	"time"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/models"
)

// AddBatch generates hash for multiple URLS and saves it along with original URL to in-memory storage and to file storage
func (r *MemoryRepository) AddBatch(ctx context.Context, urls []models.APIBatchModel, cfg *config.Config, userID string) ([]models.APIBatchResponse, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	writer, err := NewWriter(cfg.StorageFilePath)
	if err != nil {
		return nil, err
	}
	defer writer.Close()

	res := make([]models.APIBatchResponse, 0, len(urls))

	for _, urlModel := range urls {
		r.MemoryStorage[urlModel.Hash] = urlModel.OriginalURL
		urlData := models.URLData{UUID: strconv.Itoa(len(r.MemoryStorage)), ShortURL: urlModel.Hash, OriginalURL: urlModel.OriginalURL, Active: true, CreatedAt: time.Now()}
		if err = writer.WriteData(&urlData); err != nil {
			return nil, err
		}
		res = append(res, models.APIBatchResponse{CorrelationID: urlModel.CorrelationID, ShortURL: urlModel.ShortURL})
	}
	return res, nil
}
