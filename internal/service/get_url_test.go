package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/repository/mocks"
)

func TestURLShortenerService_GetURLData(t *testing.T) {
	baseURL := "https://short"
	_, err := logging.InitLogger("INFO", "dev")
	require.NoError(t, err)

	cases := []struct {
		name       string
		hash       string
		isExists   bool
		buildStubs func(s *mocks.MockRepository)
		expected   models.URLData
	}{
		{
			name:     "ok",
			hash:     "abc123",
			isExists: true,
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					GetURLData(gomock.Any(), "abc123").
					Times(1).Return(models.URLData{UUID: "abc123", OriginalURL: "https://valid1.com", ShortURL: "https://short/abc123", Active: true, CreatedAt: time.Date(2025, time.June, 10, 23, 0, 0, 0, time.UTC)}, true)

			},
			expected: models.URLData{UUID: "abc123", OriginalURL: "https://valid1.com", ShortURL: "https://short/abc123", Active: true, CreatedAt: time.Date(2025, time.June, 10, 23, 0, 0, 0, time.UTC)},
		},
		{
			name:     "not found",
			hash:     "abc123",
			isExists: false,
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					GetURLData(gomock.Any(), "abc123").
					Times(1).Return(models.URLData{}, false)

			},
			expected: models.URLData{},
		},
	}

	for i := range cases {
		tc := cases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRepository(ctrl)
			tc.buildStubs(repo)

			service := &URLShortenerService{Repo: repo, Cfg: &config.Config{BaseURL: baseURL}}
			url, exists := service.GetURLData(context.Background(), tc.hash)

			require.Equal(t, tc.isExists, exists)
			require.Equal(t, tc.expected, url)

		})
	}
}
