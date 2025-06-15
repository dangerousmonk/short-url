package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dangerousmonk/short-url/internal/logging"
)

func (h *HTTPHandler) GetURL(w http.ResponseWriter, req *http.Request) {
	hash := chi.URLParam(req, "hash")
	urlData, isExist := h.service.GetURLData(req.Context(), hash)
	if !isExist {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	if !urlData.Active {
		logging.Log.Infof("GetURL url not active | %v", urlData.ShortURL)
		w.WriteHeader(http.StatusGone)
		return
	}

	w.Header().Set("Location", urlData.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte{})
	http.Redirect(w, req, urlData.OriginalURL, http.StatusTemporaryRedirect)

}
