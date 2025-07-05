package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/repository"
	"github.com/dangerousmonk/short-url/internal/repository/mocks"
)

func TestURLShortenerService_AddURL(t *testing.T) {
	baseURL := "https://short"
	userID := "abc123"
	cfg := &config.Config{BaseURL: baseURL, MaxURLsBatchSize: 2}
	_, err := logging.InitLogger("INFO", "dev")
	require.NoError(t, err)

	cases := []struct {
		expectedError error
		buildStubs    func(s *mocks.MockRepository)
		name          string
		url           string
		shortURL      string
		wantError     bool
	}{
		{
			name:     "url invalid",
			url:      "invalid.com",
			shortURL: "",
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddShortURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			wantError:     true,
			expectedError: ErrURLInvalid,
		},
		{
			name:     "ok",
			url:      "https://valid1.com",
			shortURL: "https://short/abashd123",
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddShortURL(gomock.Any(), "https://valid1.com", gomock.Any(), cfg, userID).Times(1).
					Return("abashd123", nil)
			},
			wantError:     false,
			expectedError: nil,
		},
		{
			name:     "save error",
			url:      "https://valid1.com",
			shortURL: "",
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddShortURL(gomock.Any(), "https://valid1.com", gomock.Any(), cfg, userID).Times(1).
					Return("", errors.New("database save error"))
			},
			wantError:     true,
			expectedError: ErrSaveFailed,
		},
		{
			name:     "url already exists",
			url:      "https://valid1.com",
			shortURL: "https://short/abashd123",
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddShortURL(gomock.Any(), "https://valid1.com", gomock.Any(), cfg, userID).Times(1).
					Return("", &repository.URLExistsError{URL: "https://valid1.com", ShortURL: "abashd123", Err: "URL already exists"})
			},
			wantError:     true,
			expectedError: ErrURLExists,
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
			short, err := service.AddURL(context.Background(), tc.url, userID)

			if tc.wantError {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.shortURL, short)
			}

		})
	}
}
