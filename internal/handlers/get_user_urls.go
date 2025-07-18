package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/logging"
)

// GetUserURLs godoc
//
//	@Summary		Retreives all urls saved by user
//	@Description	GetUserURLs retreives all active urls saved by user
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Tags			API
//	@Success		200 {object}	models.APIGetUserURLsResponse
//	@Success		204
//	@Failure		401,500
//	@Router			/api/user/urls   [get]
func (h *HTTPHandler) GetUserURLs(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get(auth.UserIDHeaderName)
	w.Header().Set("Content-Type", "application/json")

	if userID == "" {
		logging.Log.Infof("GetUserURLs invalid UserId | %v", userID)
		http.Error(w, `{"error":" No UserId"}`, http.StatusUnauthorized)
		return
	}

	userURLs, err := h.service.GetUsersURLs(req.Context(), userID)
	if err != nil {
		logging.Log.Warnf("GetUserURLs error while fetching URLs | %v", err)
		http.Error(w, `{"error":" Error while fetching URLs"}`, http.StatusInternalServerError)
		return
	}

	if len(userURLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(userURLs); err != nil {
		logging.Log.Warnf("GetUserURLs error on encoding response | %v", err)
		http.Error(w, `{"error":" failed, to encode response"}`, http.StatusInternalServerError)
		return
	}

}
