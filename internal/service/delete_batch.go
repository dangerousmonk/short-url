package service

import (
	"github.com/dangerousmonk/short-url/internal/models"
)

func (s *URLShortenerService) BatchDelete(urls []string, userID string) {
	message := models.DeleteURLChannelMessage{URLs: urls, UserID: userID}
	s.DelCh <- message
}
