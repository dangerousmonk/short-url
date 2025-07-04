package service

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
)

type RepositoryMock struct {
	storage         map[string]string
	activityStorage map[string]bool
	userStorage     map[string][]string
}

func NewRepositoryMock() RepositoryMock {
	return RepositoryMock{
		storage:         map[string]string{},
		activityStorage: map[string]bool{},
		userStorage:     map[string][]string{},
	}
}

func (rm RepositoryMock) GetURLData(ctx context.Context, shortURL string) (URLData models.URLData, isExist bool) {
	fullURL, isExist := rm.storage[shortURL]
	urlData := models.URLData{OriginalURL: fullURL, Active: true}
	return urlData, isExist
}

func (rm RepositoryMock) AddShortURL(ctx context.Context, fullURL string, shortURL string, cfg *config.Config, userID string) (string, error) {
	rm.storage[shortURL] = fullURL

	userURLs := rm.userStorage[userID]
	userURLs = append(userURLs, shortURL)
	rm.userStorage[userID] = userURLs
	return shortURL, nil
}

func (rm RepositoryMock) Ping(ctx context.Context) error {
	return nil
}

func (rm RepositoryMock) AddBatch(ctx context.Context, urls []models.APIBatchModel, cfg *config.Config, userID string) ([]models.APIBatchResponse, error) {
	res := make([]models.APIBatchResponse, 0, len(urls))

	for _, urlModel := range urls {
		_, err := rm.AddShortURL(ctx, urlModel.OriginalURL, urlModel.ShortURL, cfg, userID)
		if err != nil {
			return nil, err
		}
		res = append(res, models.APIBatchResponse{CorrelationID: urlModel.CorrelationID, ShortURL: urlModel.ShortURL})
	}
	return res, nil
}

func (rm RepositoryMock) GetUsersURLs(ctx context.Context, userID, baseURL string) ([]models.APIGetUserURLsResponse, error) {
	userURLs := rm.userStorage[userID]
	if len(userURLs) == 0 {
		return nil, nil
	}
	result := make([]models.APIGetUserURLsResponse, len(userURLs))

	for _, shortURL := range userURLs {
		result = append(result, models.APIGetUserURLsResponse{
			ShortURL:    shortURL,
			OriginalURL: rm.storage[shortURL],
		})
	}
	return result, nil
}

func (rm RepositoryMock) DeleteBatch(ctx context.Context, urls []string, userID string) error {
	return nil
}

func BenchmarkShortURLService(b *testing.B) {
	_, err := logging.InitLogger("INFO", "test")
	require.NoError(b, err)

	cfg := config.Config{BaseURL: "http://localhost:8080", MaxURLsBatchSize: 50}
	service := NewShortenerService(NewRepositoryMock(), &cfg, make(chan models.DeleteURLChannelMessage))

	ctx := context.Background()
	caseNum := 10

	userID := "abcdfeg123"
	URLs := make([]string, caseNum)

	for i := 0; i < caseNum; i++ {
		URLs[i] = "http://yandex" + strconv.Itoa(i) + ".ru"
	}

	shortURL, err := service.AddURL(ctx, URLs[0], userID)
	require.NoError(b, err)

	testID := strings.Split(shortURL, "/")[3]

	b.ResetTimer()

	b.Run("AddURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err = service.AddURL(ctx, "https://github.com", userID)
			if err != nil {
				panic(err)
			}
		}
	})

	b.Run("GetURLData", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = service.GetURLData(ctx, testID)
		}
	})
}
