package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/helpers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/storage"
)

type APIShortenBatchHandler struct {
	Config  *config.Config
	Storage storage.Storage
}

func (h *APIShortenBatchHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		urls []models.APIBatchModel
		resp []models.APIBatchResponse
	)
	if err := json.NewDecoder(req.Body).Decode(&urls); err != nil {
		logging.Log.Warnf("Error on decoding body | method=%v | url=%v | err=%v", req.Method, req.URL, err)
		http.Error(w, "Error on decoding body", http.StatusInternalServerError)
		return
	}

	userID := req.Header.Get(auth.UserIDHeaderName)
	if userID == "" {
		logging.Log.Warnf("No userID in headers")
	}

	defer req.Body.Close()

	if len(urls) == 0 {
		http.Error(w, "No URLs in body", http.StatusBadRequest)
		return
	}

	for idx, url := range urls {
		if !helpers.IsURLValid(url.OriginalURL) {
			http.Error(w, fmt.Sprintf("Invalid URL: %s", url.OriginalURL), http.StatusBadRequest)
			return
		}
		hash, err := helpers.HashGenerator()
		if err != nil {
			logging.Log.Warnf("Error on generating hash | method=%v | url=%v | err=%v", req.Method, req.URL, err)
			http.Error(w, "Error on generating hash", http.StatusInternalServerError)
			return
		}
		short := h.Config.BaseURL + "/" + hash
		urls[idx].ShortURL = short
		urls[idx].Hash = hash
	}

	resp, err := h.Storage.AddBatch(req.Context(), urls, h.Config, userID)
	if err != nil {
		logging.Log.Warnf("Error on saving to storage | method=%v | url=%v | err=%v", req.Method, req.URL, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logging.Log.Warnf("Error on encoding response | %v", err)
		http.Error(w, "Error on encoding response", http.StatusInternalServerError)
		return
	}

}
