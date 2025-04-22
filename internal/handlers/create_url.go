package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/helpers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/storage"
)

type APIShortenerHandler struct {
	Config  *config.Config
	Storage storage.Storage
}

type URLShortenerHandler struct {
	Config  *config.Config
	Storage storage.Storage
}

func (h *APIShortenerHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var r models.Request
	if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
		logging.Log.Warnf("Error on decoding body | method=%v | url=%v | err=%v", req.Method, req.URL, err)
		http.Error(w, "Error on decoding body", http.StatusInternalServerError)
		return
	}

	userID := req.Header.Get(auth.UserIDHeaderName)
	if userID == "" {
		logging.Log.Warnf("No userID in headers")
	}

	defer req.Body.Close()

	if !helpers.IsURLValid(r.URL) {
		http.Error(w, "Invalid URL provided", http.StatusBadRequest)
		return
	}

	shortURL, err := h.Storage.AddShortURL(req.Context(), r.URL, h.Config, userID)
	if err != nil {
		logging.Log.Warnf("Error on inserting URL | %v", err)
		var existsErr *storage.URLExistsError
		if errors.As(err, &existsErr) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			resp := models.Response{Result: h.Config.BaseURL + "/" + existsErr.ShortURL}

			if err := json.NewEncoder(w).Encode(resp); err != nil {
				logging.Log.Warnf("Error on encoding response | %v", err)
				http.Error(w, "Error on encoding response", http.StatusInternalServerError)
				return
			}
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := models.Response{Result: h.Config.BaseURL + "/" + shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logging.Log.Warnf("Error on encoding response | %v", err)
		http.Error(w, "Error on encoding response", http.StatusInternalServerError)
		return
	}

}

func (h *URLShortenerHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	userID := req.Header.Get(auth.UserIDHeaderName)
	if userID == "" {
		logging.Log.Warnf("No userID in headers")
	}

	defer req.Body.Close()
	fullURL := string(body)

	if !helpers.IsURLValid(fullURL) {
		http.Error(w, "Invalid URL provided", http.StatusBadRequest)
		return
	}

	shortURL, err := h.Storage.AddShortURL(req.Context(), fullURL, h.Config, userID)
	if err != nil {
		logging.Log.Warnf("Error on inserting URL | %v", err)
		var existsErr *storage.URLExistsError
		if errors.As(err, &existsErr) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(h.Config.BaseURL + "/" + existsErr.ShortURL))
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logging.Log.Infof("Created url=%v", shortURL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.Config.BaseURL + "/" + shortURL))

}
