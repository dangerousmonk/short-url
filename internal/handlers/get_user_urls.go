package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/storage"
)

type APIGetUserURLsHandler struct {
	Config  *config.Config
	Storage storage.Storage
}

func (h *APIGetUserURLsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get(auth.UserIDHeaderName)
	w.Header().Set("Content-Type", "application/json")

	if userID == "" {
		logging.Log.Infof("Invalid UserId | %v", userID)
		http.Error(w, `{"error":" No UserId"}`, http.StatusUnauthorized)
		return
	}

	userURLs, err := h.Storage.GetUsersURLs(req.Context(), userID, h.Config.BaseURL)
	if err != nil {
		logging.Log.Warnf("Error while fetching URLs | %v", err)
		http.Error(w, `{"error":" Error while fetching URLs"}`, http.StatusInternalServerError)
		return
	}

	logging.Log.Infof("APIGetUserURLsHandler UserId | %v", userID)

	if len(userURLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(userURLs); err != nil {
		logging.Log.Warnf("Error on encoding response | %v", err)
		http.Error(w, `{"error":" failed, to encode response"}`, http.StatusInternalServerError)
		return
	}

}
