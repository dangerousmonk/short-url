package service

import (
	"github.com/dangerousmonk/short-url/internal/models"
)

// BatchDelete is used to send message with delete information to channel.
func (s *URLShortenerService) BatchDelete(urls []string, userID string) {
	message := models.DeleteURLChannelMessage{URLs: urls, UserID: userID}
	s.DelCh <- message
}
