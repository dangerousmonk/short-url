package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/storage/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "*.json")
	require.NoError(t, err, "Error on creating tmp file")

	defer os.Remove(tempFile.Name())

	cfg := &config.Config{
		StorageFilePath: tempFile.Name(),
	}
	mapStorage := InitMapStorage(cfg)

	expectedRows := []*Row{
		{
			UUID:        "1",
			ShortURL:    "b42b4e8f",
			OriginalURL: "https://example.com",
		},
		{
			UUID:        "2",
			ShortURL:    "78327d1f",
			OriginalURL: "https://ya.ru",
		},
	}

	for _, row := range expectedRows {
		data, err := json.Marshal(row)
		require.NoError(t, err, "Error on json.Marshal")

		data = append(data, '\n')
		_, err = tempFile.Write(data)
		require.NoError(t, err, "Error on Write to tmp file")
	}

	err = LoadFromFile(mapStorage, cfg)
	if err != nil {
		t.Fatalf("ошибка при загрузке данных из файла: %v", err)
	}
	require.NoError(t, err, "Error on LoadFromFile")

	for _, row := range expectedRows {
		actualURL, ok := mapStorage.URLdata[row.ShortURL]
		assert.True(t, ok, "Expected url missing")
		require.Equal(t, row.OriginalURL, actualURL, "Save URL differ from expected")
	}

	require.Equal(t, len(mapStorage.URLdata), len(expectedRows), "Diffrent number of rows")
}

func TestGetFullURLOk(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	m := mocks.NewMockStorage(mockCtrl)
	fullURL := "https://example.com"
	m.EXPECT().GetFullURL(context.Background(), "c2a3c895").Return(fullURL, true)

	full, exists := Storage.GetFullURL(m, context.Background(), "c2a3c895")
	require.Equal(t, full, fullURL)
	require.True(t, exists)
}

func TestGetFullURLNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	m := mocks.NewMockStorage(mockCtrl)
	m.EXPECT().GetFullURL(gomock.Any(), "fake").Return("", false)

	full, exists := m.GetFullURL(context.Background(), "fake")

	require.Empty(t, full)
	require.False(t, exists)
}

func TestAddShortURLOk(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	m := mocks.NewMockStorage(mockCtrl)
	fullURL := "https://example.com"
	hash := "cfb05b2a"
	mockCfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	m.EXPECT().AddShortURL(context.Background(), fullURL, mockCfg, gomock.Any()).Return(hash, nil)

	short, err := Storage.AddShortURL(m, context.Background(), fullURL, mockCfg, "")
	require.Equal(t, hash, short)
	require.NoError(t, err)
}

func TestAddShortURLError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	m := mocks.NewMockStorage(mockCtrl)
	fullURL := "invalid_url"
	mockCfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	m.EXPECT().AddShortURL(context.Background(), fullURL, mockCfg, gomock.Any()).Return("", errors.New("invalid URL"))

	short, err := m.AddShortURL(context.Background(), fullURL, mockCfg, "")

	require.Empty(t, short)
	require.Error(t, err)
}

func TestPingOk(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()
	m := mocks.NewMockStorage(mockCtrl)
	m.EXPECT().Ping(ctx).Return(nil)

	err := Storage.Ping(m, context.Background())
	require.NoError(t, err)
}

func TestPingFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	m := mocks.NewMockStorage(mockCtrl)
	ctx := context.Background()
	m.EXPECT().Ping(ctx).Return(errors.New("database unavailable"))

	err := m.Ping(ctx)

	require.Error(t, err)
}
