package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/go-chi/chi/v5"
)

type URLShortenerHandler struct {
	Config     *config.Config
	MapStorage *storage.MapStorage
}

type GetFullURLHandler struct {
	Config     *config.Config
	MapStorage *storage.MapStorage
}

type APIShortenerHandler struct {
	Config     *config.Config
	MapStorage *storage.MapStorage
}

func (h *URLShortenerHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	defer req.Body.Close()
	fullURL := string(body)

	if fullURL == "" {
		http.Error(w, "URL is empty", http.StatusBadRequest)
		return
	}

	shortURL, err := h.MapStorage.AddShortURL(fullURL, h.Config.StorageFilePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.Config.BaseURL + "/" + shortURL))

}

func (h *GetFullURLHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	hash := chi.URLParam(req, "hash")
	fullURL, isExist := h.MapStorage.GetFullURL(hash)
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
		logging.Log.Warnf("Error on decoding body | %v", err)
		http.Error(w, "Error on decoding body", http.StatusInternalServerError)
		return
	}

	defer req.Body.Close()

	if r.URL == "" {
		http.Error(w, "URL is empty", http.StatusBadRequest)
		return
	}

	shortURL, err := h.MapStorage.AddShortURL(r.URL, h.Config.StorageFilePath)
	if err != nil {
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
