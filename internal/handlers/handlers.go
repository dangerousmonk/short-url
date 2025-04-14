package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/helpers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/go-chi/chi/v5"
)

type URLShortenerHandler struct {
	Config  *config.Config
	Storage storage.Storage
}

type GetFullURLHandler struct {
	Config  *config.Config
	Storage storage.Storage
}

type APIShortenerHandler struct {
	Config  *config.Config
	Storage storage.Storage
}

type PingHandler struct {
	Config  *config.Config
	Storage storage.Storage
}

type APIShortenBatchHandler struct {
	Config  *config.Config
	Storage storage.Storage
}

type APIGetUserURLsHandler struct {
	Config  *config.Config
	Storage storage.Storage
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
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.Config.BaseURL + "/" + shortURL))

}

func (h *GetFullURLHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	hash := chi.URLParam(req, "hash")
	fullURL, isExist := h.Storage.GetFullURL(req.Context(), hash)
	if !isExist {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Location", fullURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte{})
	http.Redirect(w, req, fullURL, http.StatusTemporaryRedirect)

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

func (h *PingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
	defer cancel()
	err := h.Storage.Ping(ctx)
	if err != nil {
		logging.Log.Errorf("Database unreachable | %v", err)
		http.Error(w, "Database unreachable", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

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
