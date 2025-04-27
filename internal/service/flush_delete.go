package service

import (
	"context"
	"time"

	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
)

func (s *URLShortenerService) FlushDeleteMessages() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var messages []models.DeleteURLChannelMessage

	for {
		select {
		case msg := <-s.DelCh:
			messages = append(messages, msg)
		case <-ticker.C:
			if len(messages) == 0 {
				continue
			}
			for _, msg := range messages {
				err := s.Repo.DeleteBatch(context.TODO(), msg.URLs, msg.UserID)
				if err != nil {
					logging.Log.Warnf("FlushDeleteMessages error=%v", err)
					continue
				}
			}
			messages = nil
		}
	}
}
