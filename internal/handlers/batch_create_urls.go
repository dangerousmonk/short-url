package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/service"
)

// APIShortenBatch godoc
//
//	@Summary		Create batch of urls
//	@Description	APIShortenBatch is used to handle multiple urls in request
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Tags			API
//	@Param			data	body		models.APIBatchModel	true	"Request body"
//	@Success		201 {object}	models.APIBatchResponse
//	@Failure		400,401,413,500
//	@Router			/api/shorten/batch   [post]
func (h *HTTPHandler) APIShortenBatch(w http.ResponseWriter, req *http.Request) {
	var urls []models.APIBatchModel
	if err := json.NewDecoder(req.Body).Decode(&urls); err != nil {
		logging.Log.Warnf("Error on decoding body | method=%v | url=%v | err=%v", req.Method, req.URL, err)
		http.Error(w, "Error on decoding body", http.StatusBadRequest)
		return
	}

	userID := req.Header.Get(auth.UserIDHeaderName)
	if userID == "" {
		http.Error(w, "No valid cookie provided", http.StatusUnauthorized)
		return
	}

	defer req.Body.Close()

	resp, err := h.service.BatchCreate(urls, req.Context(), userID)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrTooManyURLs):
			http.Error(w, "Too many URLs", http.StatusRequestEntityTooLarge)
			return

		case errors.Is(err, service.ErrNoValidURLs):
			http.Error(w, "No valid URLs in body", http.StatusBadRequest)
			return

		case errors.Is(err, service.ErrHashFailed):
			http.Error(w, "Error on generating hash", http.StatusInternalServerError)
			return
		case errors.Is(err, service.ErrSaveBatchFailed):
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logging.Log.Warnf("Error on encoding response | %v", err)
		http.Error(w, "Error on encoding response", http.StatusInternalServerError)
		return
	}

}
