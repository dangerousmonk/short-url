package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/storage/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestPing(t *testing.T) {
	_, err := logging.InitLogger("INFO", "dev")
	require.NoError(t, err)

	cfg := config.Config{BaseURL: "http://localhost:8080"}

	testCases := []struct {
		name          string
		method        string
		buildStubs    func(s *mocks.MockStorage)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			method: http.MethodGet,
			buildStubs: func(s *mocks.MockStorage) {
				s.EXPECT().
					Ping(gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				res := w.Result()
				defer res.Body.Close()

				require.Equal(t, http.StatusOK, w.Code)
				require.Empty(t, w.Body)
			},
		},
		{
			name:   "Database returned error",
			method: http.MethodGet,
			buildStubs: func(s *mocks.MockStorage) {
				s.EXPECT().
					Ping(gomock.Any()).
					Times(1).
					Return(errors.New("Some DB error"))
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				res := w.Result()
				defer res.Body.Close()

				require.Equal(t, http.StatusInternalServerError, w.Code)
				require.NotEmpty(t, w.Body)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := mocks.NewMockStorage(ctrl)
			tc.buildStubs(s)

			req := httptest.NewRequest(tc.method, "/ping", nil)
			w := httptest.NewRecorder()

			handler := PingHandler{Config: &cfg, Storage: s}
			handler.ServeHTTP(w, req)
			require.NoError(t, err)

			tc.checkResponse(t, w)
		})
	}
}
