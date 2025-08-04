package network

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckSubnet(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		xRealIP        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid_ip_in_subnet",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ip_not_in_subnet",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "10.0.0.1",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Access forbidden\n",
		},
		{
			name:           "invalid_ip_format",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "not.an.ip",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Access forbidden\n",
		},
		{
			name:           "missing_header",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "X-Real-IP header required\n",
		},
		{
			name:           "empty_subnet",
			trustedSubnet:  "",
			xRealIP:        "192.168.1.100",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Access without trusted subnet option is forbidden\n",
		},
		{
			name:           "invalid_subnet_format",
			trustedSubnet:  "invalid.subnet",
			xRealIP:        "192.168.1.100",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Invalid trusted subnet configuration\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			cfg := &config.Config{
				TrustedSubnet: tt.trustedSubnet,
			}
			_, err := logging.InitLogger("INFO", "dev")
			require.NoError(t, err)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create request with X-Real-IP header if provided
			req := httptest.NewRequest("GET", "http://testing", nil)
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			rec := httptest.NewRecorder()

			middleware := CheckSubnet(cfg)
			middleware(handler).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}
