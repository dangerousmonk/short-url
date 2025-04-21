package handlers

import (
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/go-chi/chi/v5"
)

type GetFullURLHandler struct {
	Config  *config.Config
	Storage storage.Storage
}

func (h *GetFullURLHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	hash := chi.URLParam(req, "hash")
	urlData, isExist := h.Storage.GetURLData(req.Context(), hash)
	if !isExist {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	if !urlData.Active {
		w.WriteHeader(http.StatusGone)
		return
	}

	w.Header().Set("Location", urlData.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte{})
	http.Redirect(w, req, urlData.OriginalURL, http.StatusTemporaryRedirect)

}
