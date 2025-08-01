package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dangerousmonk/short-url/internal/logging"
)

// GetUserURLs godoc
//
//	@Summary		Retreives stats with stored URLs count and unique user_id count
//	@Description	GetStats retreives stats with stored URLs count and unique user_id count
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Tags			API
//	@Success		200 {object}	models.InternalStats
//	@Failure		500
//	@Router			/api/internal/stats   [get]
func (h *HTTPHandler) GetStats(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats, err := h.service.GetStats(req.Context())
	if err != nil {
		logging.Log.Warnf("GetStats error while fetching stats | %v", err)
		http.Error(w, `{"error":" Error while fetching stats"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		logging.Log.Warnf("GetStats error on encoding response | %v", err)
		http.Error(w, `{"error":" failed, to encode response"}`, http.StatusInternalServerError)
		return
	}

}
