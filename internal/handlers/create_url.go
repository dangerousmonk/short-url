// Package handlers that serve as main entrypoint to perform HTTP POST operations on single URL.
// Their main difference is the format of request body: APIShorten supports json body while Shorten supports plain text
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/service"
)

// APIShorten godoc
//
//	@Summary		Create single short url
//	@Description	APIShorten is used to handle single url in request
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Tags			API
//	@Param			data	body		models.Request	true	"Request body"
//	@Success		201 {object}	models.Response
//	@Failure		400,401,409,500
//	@Router			/api/shorten   [post]
func (h *HTTPHandler) APIShorten(w http.ResponseWriter, req *http.Request) {
	var r models.Request
	if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
		logging.Log.Warnf("Error on decoding body | method=%v | url=%v | err=%v", req.Method, req.URL, err)
		http.Error(w, "Error on decoding body", http.StatusBadRequest)
		return
	}

	userID := req.Header.Get(auth.UserIDHeaderName)
	if userID == "" {
		http.Error(w, "No valid cookie provided", http.StatusUnauthorized)
		return
	}

	defer req.Body.Close()

	shortURL, err := h.service.AddURL(req.Context(), r.URL, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrURLInvalid):
			http.Error(w, "Invalid URL provided", http.StatusBadRequest)
			return

		case errors.Is(err, service.ErrURLExists):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			resp := models.Response{Result: shortURL}
			if err = json.NewEncoder(w).Encode(resp); err != nil {
				logging.Log.Warnf("APIShorten error on encoding response | %v", err)
				http.Error(w, "Error on encoding response", http.StatusInternalServerError)
				return
			}
			return

		case errors.Is(err, service.ErrSaveFailed):
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		default:
			logging.Log.Warnf("APIShortenerHandler unknown err:%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
	resp := models.Response{Result: shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logging.Log.Warnf("Error on encoding response | %v", err)
		http.Error(w, "Error on encoding response", http.StatusInternalServerError)
		return
	}

}

// Shorten godoc
//
//	@Summary		Used to short single URL
//	@Description	Used to short single URL provided in request body
//	@Accept			plain
//	@Produce		plain
//	@Param			data	body		string	true	"Request body"
//	@Tags			URL
//	@Success		201
//	@Failure		400,401,409,500
//	@Router			/   [post]
func (h *HTTPHandler) Shorten(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	userID := req.Header.Get(auth.UserIDHeaderName)
	if userID == "" {
		http.Error(w, "No valid cookie provided", http.StatusUnauthorized)
		return
	}

	defer req.Body.Close()
	fullURL := string(body)

	logging.Log.Infof("adding url from body=%v", fullURL)
	shortURL, err := h.service.AddURL(req.Context(), fullURL, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrURLInvalid):
			http.Error(w, "Invalid URL provided", http.StatusBadRequest)
			return

		case errors.Is(err, service.ErrURLExists):
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(shortURL))
			return

		case errors.Is(err, service.ErrSaveFailed):
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		default:
			logging.Log.Warnf("Shorten unknown err:%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
	logging.Log.Infof("Created url=%v for userId=%v", shortURL, userID)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))

}
