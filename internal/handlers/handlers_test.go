package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestURLShortenerHandler(t *testing.T) {
	type expected struct {
		statusCode  int
		contentType string
	}
	cfg := config.Config{BaseURL: "http://localhost:8080"}
	storage := &storage.MapStorage{
		URLdata: make(map[string]string),
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

			shortenHandler := URLShortenerHandler{Config: &cfg, MapStorage: storage}
			shortenHandler.ServeHTTP(w, req)

			result := w.Result()
			defer result.Body.Close()

			require.Equal(t, test.expected.statusCode, w.Code, "Код ответ не совпадает с ожидаемым")

			if test.expected.statusCode == http.StatusCreated {
				require.Equal(t, test.expected.contentType, result.Header.Get("Content-Type"), "Content-Type не совпадает с ожидаемым")

				body, err := io.ReadAll(result.Body)
				require.NoError(t, err)

				shortURL := string(body)
				hash := strings.TrimPrefix(shortURL, cfg.BaseURL+"/")
				fullURL, isExist := storage.GetFullURL(hash)

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
	cfg := config.Config{
		BaseURL:    "http://localhost:8080",
		ServerAddr: "http://localhost:8080",
	}
	storage := storage.NewMapStorage()
	storage.URLdata["dfccf368"] = "https://example.com"
	storage.URLdata["65f7ae83"] = "https://www.google.com"

	getURLhandler := GetFullURLHandler{Config: &cfg, MapStorage: storage}

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
			name:     "Check not exist URL",
			method:   http.MethodGet,
			hash:     "azc1f3fp",
			expected: expected{statusCode: http.StatusNotFound, location: ""},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx := chi.NewRouteContext()
			req := httptest.NewRequest(test.method, cfg.ServerAddr+"/"+test.hash, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
			ctx.URLParams.Add("hash", test.hash)

			getURLhandler.ServeHTTP(w, req)

			result := w.Result()
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

func TestAPIShortenerHandler(t *testing.T) {
	type expected struct {
		statusCode  int
		contentType string
	}
	cfg := config.Config{BaseURL: "http://localhost:8080"}
	storage := &storage.MapStorage{
		URLdata: make(map[string]string),
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
			body:     `{"url":"https://example.com"}`,
			expected: expected{statusCode: http.StatusCreated, contentType: "application/json"},
		},
		{
			name:     "Check empty body",
			method:   http.MethodPost,
			body:     `{}`,
			expected: expected{statusCode: http.StatusBadRequest, contentType: ""},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var req models.Request
			err := json.NewDecoder(strings.NewReader(test.body)).Decode(&req)
			if err != nil {
				t.Errorf("error not nil on decoding | %v", err)
			}

			testReq := httptest.NewRequest(test.method, "/api/shorten", strings.NewReader(test.body))
			w := httptest.NewRecorder()

			shortenHandler := APIShortenerHandler{Config: &cfg, MapStorage: storage}
			shortenHandler.ServeHTTP(w, testReq)

			result := w.Result()
			defer result.Body.Close()

			require.Equal(t, test.expected.statusCode, w.Code, "Код ответ не совпадает с ожидаемым")

			if test.expected.statusCode == http.StatusCreated {
				require.Equal(t, test.expected.contentType, result.Header.Get("Content-Type"), "Content-Type не совпадает с ожидаемым")
				response := models.Response{}
				json.NewDecoder(result.Body).Decode(&response)

				hash := strings.TrimPrefix(response.Result, cfg.BaseURL+"/")
				fullURL, isExist := storage.GetFullURL(hash)

				require.Equal(t, req.URL, fullURL, "Сохраненый URL не совпадает с ожидаемым")
				require.True(t, isExist, "Флаг сохранения URL не совпадает с ожидаемым")

			}
		})
	}
}
