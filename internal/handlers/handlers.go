package handlers

import (
	"io"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/go-chi/chi/v5"
)

func URLShortenerHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

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

	shortURL := storage.AppStorage.AddShortURL(fullURL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(config.Cfg.BaseURL + "/" + shortURL))

}

func GetFullURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	hash := chi.URLParam(r, "hash")
	fullURL, isExist := storage.AppStorage.GetFullURL(hash)
	if !isExist {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Location", fullURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte{})
}
