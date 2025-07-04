package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dangerousmonk/short-url/internal/logging"
)

// GetURL godoc
//
//	@Summary		Redirects to the original URL
//	@Description	GetURL redirects to the original URL by using short url
//	@Accept			plain
//	@Produce		plain
//	@Tags			URL
//	@Success		307
//	@Failure		404,410
//	@Router			/{hash}   [get]
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
