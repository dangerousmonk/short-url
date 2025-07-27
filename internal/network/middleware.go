package network

import (
	"net"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/logging"
)

// CheckSubnet is middleware that validates if source address matches the trusted subnet
func CheckSubnet(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if cfg.TrustedSubnet == "" {
				http.Error(w, "Access without trusted subnet option is forbidden", http.StatusForbidden)
				return
			}
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				http.Error(w, "X-Real-IP header required", http.StatusForbidden)
				return
			}
			_, subnet, err := net.ParseCIDR(cfg.TrustedSubnet)
			if err != nil {
				logging.Log.Warnf("CheckSubnet ParseCIDR error %v", err)
				http.Error(w, "Invalid trusted subnet configuration", http.StatusInternalServerError)
				return
			}
			ip := net.ParseIP(realIP)
			if ip == nil || !subnet.Contains(ip) {
				http.Error(w, "Access forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
