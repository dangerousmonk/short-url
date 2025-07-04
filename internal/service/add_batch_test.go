package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/repository/mocks"
)

func TestURLShortenerService_BatchCreate(t *testing.T) {
	baseURL := "https://short"
	userID := "abc123"
	cfg := &config.Config{BaseURL: baseURL, MaxURLsBatchSize: 2}
	_, err := logging.InitLogger("INFO", "dev")
	require.NoError(t, err)

	cases := []struct {
		name          string
		input         []models.APIBatchModel
		expected      []models.APIBatchResponse
		buildStubs    func(s *mocks.MockRepository)
		wantError     bool
		expectedError error
	}{
		{
			name: "too many urls",
			input: []models.APIBatchModel{
				{CorrelationID: "1", OriginalURL: "https://valid1.com"},
				{CorrelationID: "2", OriginalURL: "https://valid2.com"},
				{CorrelationID: "3", OriginalURL: "https://valid3.com"},
			},
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddBatch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expected:      nil,
			wantError:     true,
			expectedError: ErrTooManyURLs,
		},
		{
			name: "only invalid urls",
			input: []models.APIBatchModel{
				{CorrelationID: "1", OriginalURL: ""},
				{CorrelationID: "2", OriginalURL: "noturl.com"},
			},
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddBatch(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any()).
					Times(0)
			},
			expected:      nil,
			wantError:     true,
			expectedError: ErrNoValidURLs,
		},
		{
			name: "ok",
			input: []models.APIBatchModel{
				{CorrelationID: "1", OriginalURL: "https://valid1.com"},
				{CorrelationID: "2", OriginalURL: "https://valid2.com"},
			},
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddBatch(
						gomock.Any(),
						gomock.Any(),
						cfg,
						userID).
					Times(1).
					Return(
						[]models.APIBatchResponse{
							{CorrelationID: "1", ShortURL: "https://short/abc123"},
							{CorrelationID: "2", ShortURL: "https://short/def456"},
						},
						nil)

			},
			expected: []models.APIBatchResponse{
				{CorrelationID: "1", ShortURL: "https://short/abc123"},
				{CorrelationID: "2", ShortURL: "https://short/def456"}},
			wantError: false,
		},
		{
			name: "repo save error",
			input: []models.APIBatchModel{
				{CorrelationID: "1", OriginalURL: "https://valid1.com"},
				{CorrelationID: "2", OriginalURL: "https://valid2.com"},
			},
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddBatch(
						gomock.Any(),
						gomock.Any(),
						cfg,
						userID).
					Times(1).
					Return(
						nil,
						errors.New("failed to save urls"))

			},
			expected:      nil,
			wantError:     true,
			expectedError: ErrSaveBatchFailed,
		},
	}

	for i := range cases {
		tc := cases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRepository(ctrl)
			tc.buildStubs(repo)

			service := &URLShortenerService{Repo: repo, Cfg: cfg}
			urls, err := service.BatchCreate(context.Background(), tc.input, userID)

			if tc.wantError {
				require.Error(t, err)
				require.Empty(t, urls)
				require.EqualError(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, urls)
			}

		})
	}
}
