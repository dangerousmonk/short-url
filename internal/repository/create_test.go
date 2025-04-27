package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/repository/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAddShortURLOk(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	m := mocks.NewMockRepository(mockCtrl)
	fullURL := "https://example.com"
	hash := "cfb05b2a"
	mockCfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	m.EXPECT().AddShortURL(context.Background(), fullURL, mockCfg, gomock.Any()).Return(hash, nil)

	short, err := m.AddShortURL(context.Background(), fullURL, mockCfg, "")
	require.Equal(t, hash, short)
	require.NoError(t, err)
}

func TestAddShortURLError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	m := mocks.NewMockRepository(mockCtrl)
	fullURL := "invalid_url"
	mockCfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	m.EXPECT().AddShortURL(context.Background(), fullURL, mockCfg, gomock.Any()).Return("", errors.New("invalid URL"))

	short, err := m.AddShortURL(context.Background(), fullURL, mockCfg, "")

	require.Empty(t, short)
	require.Error(t, err)
}
