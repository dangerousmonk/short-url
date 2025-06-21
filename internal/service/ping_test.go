package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/dangerousmonk/short-url/internal/repository/mocks"
)

func TestURLShortenerService_Ping(t *testing.T) {
	cases := []struct {
		name       string
		buildStubs func(s *mocks.MockRepository)
		wantError  bool
	}{
		{
			name:      "ping ok",
			wantError: false,
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					Ping(gomock.Any()).
					Times(1).Return(nil)
			},
		},
		{
			name:      "ping error",
			wantError: true,
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					Ping(gomock.Any()).
					Times(1).Return(errors.New("connection refused"))
			},
		},
	}

	for i := range cases {
		tc := cases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRepository(ctrl)
			tc.buildStubs(repo)

			service := &URLShortenerService{Repo: repo}
			err := service.Ping(context.Background())

			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

		})
	}
}
