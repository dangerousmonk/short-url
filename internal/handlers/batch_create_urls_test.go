package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/repository/mocks"
	"github.com/dangerousmonk/short-url/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func requireBodyMatch(t *testing.T, body *bytes.Buffer, resp []models.APIBatchResponse) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var urls []models.APIBatchResponse
	err = json.Unmarshal(data, &urls)
	require.NoError(t, err)
	require.Equal(t, resp, urls)
}

func TestAPIShortenBatch(t *testing.T) {
	_, err := logging.InitLogger("INFO", "dev")
	require.NoError(t, err)

	cfg := config.Config{BaseURL: "http://localhost:8080", MaxURLsBatchSize: 50}

	urls := []models.APIBatchResponse{
		{
			CorrelationID: "cea4eb67",
			ShortURL:      "http://localhost:8080/3daf1e25",
		},
		{
			CorrelationID: "3b936c58",
			ShortURL:      "http://localhost:8080/cfb05b2a",
		},
	}

	testCases := []struct {
		name          string
		method        string
		body          string
		expectedCode  int
		buildStubs    func(s *mocks.MockRepository)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
		userHeader    string
	}{
		{
			name:         "Request without headers",
			method:       http.MethodPost,
			body:         `[{"correlation_id": "cea4eb67","original_url": "https://stackoverflow.com"}, {"correlation_id": "3b936c58","original_url": "https://github.com"}]`,
			expectedCode: http.StatusUnauthorized,
			userHeader:   "",
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddBatch(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name:         "OK",
			method:       http.MethodPost,
			body:         `[{"correlation_id": "cea4eb67","original_url": "https://stackoverflow.com"}, {"correlation_id": "3b936c58","original_url": "https://github.com"}]`,
			expectedCode: http.StatusCreated,
			userHeader:   "b714f6f3232240c48e56029c3e65730d",
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddBatch(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(urls, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				res := w.Result()
				defer res.Body.Close()

				require.Equal(t, http.StatusCreated, w.Code)
				require.NotEmpty(t, w.Body)
				require.Equal(t, "application/json", res.Header.Get("Content-Type"))
				requireBodyMatch(t, w.Body, urls)
			},
		},
		{
			name:         "Empty body",
			method:       http.MethodPost,
			body:         `[]`,
			expectedCode: http.StatusBadRequest,
			userHeader:   "b714f6f3232240c48e56029c3e65730d",
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddBatch(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
		{
			name:         "Bad url",
			method:       http.MethodPost,
			body:         `[{"correlation_id": "cea4eb67","original_url": ""}, {"correlation_id": "3b936c58","original_url": "https://github.com"}]`,
			expectedCode: http.StatusCreated,
			userHeader:   "b714f6f3232240c48e56029c3e65730d",
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddBatch(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return([]models.APIBatchResponse{
						{
							CorrelationID: "3b936c58",
							ShortURL:      "http://localhost:8080/cfb05b2a",
						},
					}, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				res := w.Result()
				defer res.Body.Close()

				require.Equal(t, http.StatusCreated, w.Code)
				require.NotEmpty(t, w.Body)
				require.Equal(t, "application/json", res.Header.Get("Content-Type"))

				requireBodyMatch(t, w.Body, []models.APIBatchResponse{
					{
						CorrelationID: "3b936c58",
						ShortURL:      "http://localhost:8080/cfb05b2a",
					},
				})
			},
		},
		{
			name:         "Database issue",
			method:       http.MethodPost,
			body:         `[{"correlation_id": "cea4eb67","original_url": "https://stackoverflow.com"}, {"correlation_id": "3b936c58","original_url": "https://github.com"}]`,
			expectedCode: http.StatusCreated,
			userHeader:   "b714f6f3232240c48e56029c3e65730d",
			buildStubs: func(r *mocks.MockRepository) {
				r.EXPECT().
					AddBatch(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, w.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRepository(ctrl)
			tc.buildStubs(repo)

			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(tc.body))
			req.Header.Set(auth.UserIDHeaderName, tc.userHeader)
			w := httptest.NewRecorder()

			service := service.URLShortenerService{Repo: repo, Cfg: &cfg, DelCh: make(chan models.DeleteURLChannelMessage)}

			handler := NewHandler(service)
			handler.APIShortenBatch(w, req)
			require.NoError(t, err)

			tc.checkResponse(t, w)
		})
	}

}
