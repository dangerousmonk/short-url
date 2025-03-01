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
	defer req.Body.Close()
	fullURL := string(body)
	if fullURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := storage.AppStorage.AddShortURL(fullURL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/" + shortURL))

}

func GetFullURLHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer req.Body.Close()

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
