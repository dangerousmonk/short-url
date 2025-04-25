package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/storage"
)

type PingHandler struct {
	Config  *config.Config
	Storage storage.Storage
}

func (h *PingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
	defer cancel()
	err := h.Storage.Ping(ctx)
	if err != nil {
		logging.Log.Errorf("PingHandler database unreachable | %v", err)
		http.Error(w, "Database unreachable", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}
