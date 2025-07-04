package repository

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/repository/mocks"
)

func TestGetFullURLOk(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	m := mocks.NewMockRepository(mockCtrl)
	fullURL := "https://example.com"
	m.EXPECT().GetURLData(context.Background(), "c2a3c895").Return(models.URLData{OriginalURL: fullURL}, true)

	urlData, exists := m.GetURLData(context.Background(), "c2a3c895")
	require.Equal(t, urlData.OriginalURL, fullURL)
	require.True(t, exists)
}

func TestGetFullURLNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	m := mocks.NewMockRepository(mockCtrl)
	m.EXPECT().GetURLData(gomock.Any(), "fake").Return(models.URLData{}, false)

	urlData, exists := m.GetURLData(context.Background(), "fake")

	require.Empty(t, urlData.OriginalURL)
	require.False(t, exists)
}
