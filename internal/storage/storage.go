package storage

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/helpers"
	"github.com/dangerousmonk/short-url/internal/models"
)

type Row struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

//go:generate mockgen -package mocks -source storage.go -destination ./mocks/mock_storage.go Storage
type Storage interface {
	// GetURLData retrieves the original URL data by hash
	GetURLData(ctx context.Context, shortURL string) (URLData models.URLData, isExist bool)
	// AddShortURL generates hash for provided URL and saves it along with original URL to internal storage
	AddShortURL(ctx context.Context, fullURL string, cfg *config.Config, userID string) (shortURL string, err error)
	// Ping checks whether internal storage is up and running
	Ping(ctx context.Context) error
	// AddBatch generates hash for multiple URLS and saves it along with original URL to internal storage
	AddBatch(ctx context.Context, urls []models.APIBatchModel, cfg *config.Config, userID string) ([]models.APIBatchResponse, error)
	// GetUsersURLs retrieves all saved URLs by specific user
	GetUsersURLs(ctx context.Context, userID, baseURL string) ([]models.APIGetUserURLsResponse, error)
	// DeleteBatch marks multiple records as not active
	DeleteBatch(ctx context.Context, urls []models.DeleteURLChannelMessage) error
}

type MapStorage struct {
	MemoryStorage map[string]string
	mutex         sync.RWMutex
	cfg           *config.Config
}

func InitMapStorage(cfg *config.Config) *MapStorage {
	return &MapStorage{
		MemoryStorage: make(map[string]string),
		cfg:           cfg,
	}
}

func (s *MapStorage) GetURLData(ctx context.Context, shortURL string) (urlData models.URLData, isExist bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	fullURL, isExist := s.MemoryStorage[shortURL]
	urlData = models.URLData{OriginalURL: fullURL, Active: true}
	return urlData, isExist
}

func (s *MapStorage) AddShortURL(ctx context.Context, fullURL string, cfg *config.Config, userID string) (shortURL string, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for {
		shortURL, err = helpers.HashGenerator()
		if err != nil {
			return "", err
		}

		if _, exists := s.MemoryStorage[shortURL]; !exists {
			break
		}
	}

	s.MemoryStorage[shortURL] = fullURL
	urlData := Row{UUID: strconv.Itoa(len(s.MemoryStorage)), ShortURL: shortURL, OriginalURL: fullURL}

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

func (s *MapStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MapStorage) AddBatch(ctx context.Context, urls []models.APIBatchModel, cfg *config.Config, userID string) ([]models.APIBatchResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	writer, err := NewWriter(cfg.StorageFilePath)
	if err != nil {
		return nil, err
	}
	defer writer.Close()

	res := make([]models.APIBatchResponse, 0, len(urls))

	for _, urlModel := range urls {
		s.MemoryStorage[urlModel.Hash] = urlModel.OriginalURL
		urlData := Row{UUID: strconv.Itoa(len(s.MemoryStorage)), ShortURL: urlModel.Hash, OriginalURL: urlModel.OriginalURL}
		if err = writer.WriteData(&urlData); err != nil {
			return nil, err
		}
		res = append(res, models.APIBatchResponse{CorrelationID: urlModel.CorrelationID, ShortURL: urlModel.ShortURL})
	}
	return res, nil
}

func LoadFromFile(s *MapStorage, cfg *config.Config) error {
	reader, err := NewFileReader(cfg.StorageFilePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err = reader.ReadData(s)
	if err != nil {
		return err
	}
	return nil
}

func (s *MapStorage) GetUsersURLs(ctx context.Context, userID, baseURL string) ([]models.APIGetUserURLsResponse, error) {
	return nil, errors.New("mapStorage doesnt support GetUsersURLs method")
}

func (s *MapStorage) DeleteBatch(ctx context.Context, urls []models.DeleteURLChannelMessage) error {
	return errors.New("mapStorage doesnt support DeleteBatch method")
}
