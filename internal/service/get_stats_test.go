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

func TestGetStats(t *testing.T) {
	baseURL := "https://short"
	_, err := logging.InitLogger("INFO", "dev")
	require.NoError(t, err)

	cases := []struct {
		buildStubs func(r *mocks.MockRepository)
		name       string
		expected   models.InternalStats
		withError  bool
	}{
		{
			name: "ok",
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					GetStats(gomock.Any()).
					Times(1).Return(models.InternalStats{URLs: 3, Users: 2}, nil)

			},
			expected:  models.InternalStats{URLs: 3, Users: 2},
			withError: false,
		},
		{
			name: "repo error",
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					GetStats(gomock.Any()).
					Times(1).Return(models.InternalStats{}, errors.New("No rows"))

			},
			expected:  models.InternalStats{},
			withError: true,
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
			stats, err := service.GetStats(context.Background())

			if tc.withError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expected, stats)

		})
	}
}
