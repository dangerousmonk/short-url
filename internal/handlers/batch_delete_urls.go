package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
)

// APIDeleteBatch godoc
//
//	@Summary		Deletes batch of urls from storage by user
//	@Description	APIDeleteBatch is used to set active flag=false for multiple url records for user
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Tags			API
//	@Success		202
//	@Failure		400,401
//	@Router			/api/user/urls   [delete]
func (h *HTTPHandler) APIDeleteBatch(w http.ResponseWriter, req *http.Request) {
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
		http.Error(w, `{"error":" Cannot decode JSON body"}`, http.StatusBadRequest)
		return
	}
	if len(urls) == 0 {
		http.Error(w, `{"error":" No URL in body"}`, http.StatusBadRequest)
		return
	}

	message := models.DeleteURLChannelMessage{URLs: urls, UserID: userID}
	h.service.DelCh <- message

	w.WriteHeader(http.StatusAccepted)
}
