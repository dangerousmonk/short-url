package handlers

import (
	"net/http"

	"github.com/dangerousmonk/short-url/internal/logging"
)

// Ping godoc
// @Summary      DB Healthcheck
// @Description  Checks if DB is alive
// @Tags         Ping
// @Success      200
// @Failure      500
// @Router       /ping [get]
func (h *HTTPHandler) Ping(w http.ResponseWriter, req *http.Request) {
	err := h.service.Ping(req.Context())
	if err != nil {
		logging.Log.Errorf("Ping database unreachable | %v", err)
		http.Error(w, "Database unreachable", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}
