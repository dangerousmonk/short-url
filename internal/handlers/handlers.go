package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/dangerousmonk/short-url/internal/storage"
)

func URLShortenerHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fullURL := string(body)
	if fullURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortUrl := storage.AppStorage.AddShortURL(fullURL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/" + shortUrl))

}

func GetFullURLHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	path := req.URL.Path
	parts := strings.Split(path, "/")
	id := parts[len(parts)-1]
	fullURL, isExist := storage.AppStorage.GetFullURL(id)
	if !isExist {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Location", fullURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte{})
}
