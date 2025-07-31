package service

import (
	"context"

	"github.com/dangerousmonk/short-url/internal/models"
)

// BatchDelete is used to send message with delete information to channel.
func (s *URLShortenerService) DeleteBatch(ctx context.Context, urls []string, userID string) error {
	message := models.DeleteURLChannelMessage{URLs: urls, UserID: userID}
	s.DelCh <- message
	return nil
}
