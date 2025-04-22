package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/storage"
)

type APIDeleteUserURLsHandler struct {
	Config  *config.Config
	Storage storage.Storage
	DoneCh  chan models.DeleteURLChannelMessage
}

func (h *APIDeleteUserURLsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get(auth.UserIDHeaderName)
	if userID == "" {
		logging.Log.Infof("APIDeleteUserURLsHandler | Invalid UserId | %v", userID)
		http.Error(w, `{"error":" No UserId"}`, http.StatusUnauthorized)
		return
	}

	defer req.Body.Close()
	var urls []string

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&urls); err != nil {
		logging.Log.Warnf("Error on decoding body | %v", err)
		http.Error(w, `{"error":" Cannot decode JSON body"}`, http.StatusInternalServerError)
		return
	}
	if len(urls) == 0 {
		http.Error(w, `{"error":" No URL in body"}`, http.StatusBadRequest)
		return
	}

	logging.Log.Infof("APIDeleteUserURLsHandler userId=%v urls to delete | %v", userID, urls)

	message := models.DeleteURLChannelMessage{URLs: urls, UserID: userID}
	h.DoneCh <- message

	w.WriteHeader(http.StatusAccepted)
}
