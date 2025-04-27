package memory

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/models"
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
	memoryRepo := NewMemoryRepository(cfg)

	expectedRows := []*models.URLData{
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

	err = LoadFromFile(memoryRepo, cfg)
	if err != nil {
		t.Fatalf("ошибка при загрузке данных из файла: %v", err)
	}
	require.NoError(t, err, "Error on LoadFromFile")

	for _, row := range expectedRows {
		actualURL, ok := memoryRepo.MemoryStorage[row.ShortURL]
		assert.True(t, ok, "Expected url missing")
		require.Equal(t, row.OriginalURL, actualURL, "Save URL differ from expected")
	}

	require.Equal(t, len(memoryRepo.MemoryStorage), len(expectedRows), "Diffrent number of rows")
}
