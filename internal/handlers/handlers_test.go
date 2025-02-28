package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestURLShortenerHandler(t *testing.T) {
	type expected struct {
		statusCode  int
		contentType string
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
			request := httptest.NewRequest(test.method, "/", strings.NewReader(test.body))
			w := httptest.NewRecorder()
			URLShortenerHandler(w, request)

			result := w.Result()
			defer result.Body.Close()

			require.Equal(t, test.expected.statusCode, w.Code, "Код ответ не совпадает с ожидаемым")

			if test.expected.statusCode == http.StatusOK {
				require.Equal(t, test.expected.contentType, result.Header.Get("Content-Type"), "Content-Type не совпадает с ожидаемым")
				_, err := io.ReadAll(result.Body)
				require.NoError(t, err)
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

	cases := []struct {
		name     string
		method   string
		path     string
		expected expected
	}{
		{
			name:     "Check happy case",
			method:   http.MethodGet,
			path:     "/dfccf368",
			expected: expected{statusCode: http.StatusTemporaryRedirect, location: "https://example.com"},
		},
		{
			name:     "Check POST not allowed",
			method:   http.MethodPost,
			path:     "/dfccf368",
			expected: expected{statusCode: http.StatusMethodNotAllowed, location: ""},
		},
		{
			name:     "Check not exist URL",
			method:   http.MethodGet,
			path:     "/azc1f3fp",
			expected: expected{statusCode: http.StatusNotFound, location: ""},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			w := httptest.NewRecorder()
			GetFullURLHandler(w, request)

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
