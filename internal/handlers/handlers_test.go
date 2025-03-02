package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestURLShortenerHandler(t *testing.T) {
	type expected struct {
		statusCode  int
		contentType string
	}
	config.Cfg = &config.Config{
		BaseURL: "http://localhost:8080",
	}

	cases := []struct {
		name     string
		method   string
		body     string
		expected expected
	}{
		{
			name:     "Check happy case",
			method:   http.MethodPost,
			body:     "https://example.com",
			expected: expected{statusCode: http.StatusCreated, contentType: "text/plain"},
		},
		{
			name:     "Check GET not allowed",
			method:   http.MethodGet,
			body:     "",
			expected: expected{statusCode: http.StatusMethodNotAllowed, contentType: ""},
		},
		{
			name:     "Check empty body",
			method:   http.MethodPost,
			body:     "",
			expected: expected{statusCode: http.StatusBadRequest, contentType: ""},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, "/", strings.NewReader(test.body))
			w := httptest.NewRecorder()
			URLShortenerHandler(w, req)

			result := w.Result()
			defer result.Body.Close()

			require.Equal(t, test.expected.statusCode, w.Code, "Код ответ не совпадает с ожидаемым")

			if test.expected.statusCode == http.StatusCreated {
				require.Equal(t, test.expected.contentType, result.Header.Get("Content-Type"), "Content-Type не совпадает с ожидаемым")

				body, err := io.ReadAll(result.Body)
				require.NoError(t, err)

				shortURL := string(body)
				hash := strings.TrimPrefix(shortURL, config.Cfg.BaseURL+"/")
				fullURL, isExist := storage.AppStorage.GetFullURL(hash)

				require.Equal(t, test.body, fullURL, "Сохраненый URL не совпадает с ожидаемым")
				require.True(t, isExist, "Флаг сохранения URL не совпадает с ожидаемым")

				err = result.Body.Close()
				require.NoError(t, err)
			}
		})
	}
}

func TestGetFullURLHandler(t *testing.T) {
	type expected struct {
		statusCode int
		location   string
	}

	storage.AppStorage.URLdata["dfccf368"] = "https://example.com"
	storage.AppStorage.URLdata["65f7ae83"] = "https://www.google.com"

	handler := http.HandlerFunc(GetFullURLHandler)
	s := httptest.NewServer(handler)
	defer s.Close()

	cases := []struct {
		name     string
		method   string
		hash     string
		expected expected
	}{
		{
			name:     "Check happy case",
			method:   http.MethodGet,
			hash:     "dfccf368",
			expected: expected{statusCode: http.StatusTemporaryRedirect, location: "https://example.com"},
		},
		{
			name:     "Check POST not allowed",
			method:   http.MethodPost,
			hash:     "dfccf368",
			expected: expected{statusCode: http.StatusMethodNotAllowed, location: ""},
		},
		{
			name:     "Check not exist URL",
			method:   http.MethodGet,
			hash:     "azc1f3fp",
			expected: expected{statusCode: http.StatusNotFound, location: ""},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, s.URL+"/"+test.hash, nil)
			ctx := chi.NewRouteContext()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
			ctx.URLParams.Add("hash", test.hash)

			w := httptest.NewRecorder()
			handler(w, req)

			result := w.Result()
			defer result.Body.Close()

			require.Equal(t, test.expected.statusCode, w.Code, "Код ответ не совпадает с ожидаемым")

			if test.expected.statusCode == http.StatusTemporaryRedirect {
				require.Equal(t, test.expected.location, result.Header.Get("Location"), "Location не совпадает с ожидаемым")
				_, err := io.ReadAll(result.Body)
				require.NoError(t, err)
				err = result.Body.Close()
				require.NoError(t, err)
			}
		})
	}
}
