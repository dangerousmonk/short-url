package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/dangerousmonk/short-url/internal/repository/mocks"
)

func TestPingOk(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()
	m := mocks.NewMockRepository(mockCtrl)
	m.EXPECT().Ping(ctx).Return(nil)

	err := m.Ping(context.Background())
	require.NoError(t, err)
}

func TestPingFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	m := mocks.NewMockRepository(mockCtrl)
	ctx := context.Background()
	m.EXPECT().Ping(ctx).Return(errors.New("database unavailable"))

	err := m.Ping(ctx)

	require.Error(t, err)
}
