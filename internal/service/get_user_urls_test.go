package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/repository/mocks"
)

func TestURLShortenerService_GetUsersURLs(t *testing.T) {
	baseURL := "https://short"

	cases := []struct {
		buildStubs func(s *mocks.MockRepository)
		name       string
		userID     string
		expected   []models.APIGetUserURLsResponse
		wantError  bool
	}{
		{
			name:      "ok",
			userID:    "abc123",
			wantError: false,
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					GetUsersURLs(gomock.Any(), "abc123", baseURL).
					Times(1).Return([]models.APIGetUserURLsResponse{{
					OriginalURL: "https://example.com",
					ShortURL:    "https://short/foo123",
					Hash:        "foo123",
				}}, nil)

			},
			expected: []models.APIGetUserURLsResponse{{
				OriginalURL: "https://example.com",
				ShortURL:    "https://short/foo123",
				Hash:        "foo123",
			}},
		},
		{
			name:      "no urls",
			userID:    "abc123",
			wantError: false,
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					GetUsersURLs(gomock.Any(), "abc123", baseURL).
					Times(1).Return([]models.APIGetUserURLsResponse{}, nil)

			},
			expected: []models.APIGetUserURLsResponse{},
		},
		{
			name:      "database error",
			userID:    "abc123",
			wantError: true,
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					GetUsersURLs(gomock.Any(), "abc123", baseURL).
					Times(1).Return([]models.APIGetUserURLsResponse{}, errors.New("database error"))

			},
			expected: nil,
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
			urls, err := service.GetUsersURLs(context.Background(), tc.userID)

			if tc.wantError {
				require.Error(t, err)
				require.Empty(t, urls)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, urls)
			}

		})
	}
}
