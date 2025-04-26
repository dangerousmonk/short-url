package handlers

import (
	"encoding/json"
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
		urls      []models.APIBatchModel
		validURLs []models.APIBatchModel
		resp      []models.APIBatchResponse
	)
	if err := json.NewDecoder(req.Body).Decode(&urls); err != nil {
		logging.Log.Warnf("Error on decoding body | method=%v | url=%v | err=%v", req.Method, req.URL, err)
		http.Error(w, "Error on decoding body", http.StatusBadRequest)
		return
	}

	userID := req.Header.Get(auth.UserIDHeaderName)
	if userID == "" {
		logging.Log.Warnf("No userID in headers")
	}

	defer req.Body.Close()

	if len(urls) > h.Config.MaxURLsBatchSize {
		logging.Log.Warnf("APIShortenBatchHandler too many URLs=%v, allowed size=%v", len(urls), h.Config.MaxURLsBatchSize)
		http.Error(w, "Too many URLs", http.StatusRequestEntityTooLarge)
		return
	}

	for _, url := range urls {
		if helpers.IsURLValid(url.OriginalURL) {
			validURLs = append(validURLs, url)
		}
	}

	if len(validURLs) == 0 {
		http.Error(w, "No valid URLs in body", http.StatusBadRequest)
		return
	}

	for idx := range validURLs {
		hash, err := helpers.HashGenerator()
		if err != nil {
			logging.Log.Warnf("Error on generating hash | method=%v | url=%v | err=%v", req.Method, req.URL, err)
			http.Error(w, "Error on generating hash", http.StatusInternalServerError)
			return
		}
		short := h.Config.BaseURL + "/" + hash
		validURLs[idx].ShortURL = short
		validURLs[idx].Hash = hash
	}

	resp, err := h.Storage.AddBatch(req.Context(), validURLs, h.Config, userID)
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
