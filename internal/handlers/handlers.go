package handlers

import (
	"io"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
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

	shortURL := h.MapStorage.AddShortURL(fullURL)
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
	w.WriteHeader(http.StatusFound)
	w.Write([]byte{})

}
