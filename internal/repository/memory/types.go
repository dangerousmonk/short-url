package memory

import (
	"sync"

	"github.com/dangerousmonk/short-url/cmd/config"
)

type MemoryRepository struct {
	MemoryStorage map[string]string
	mutex         sync.RWMutex
	cfg           *config.Config
}

func NewMemoryRepository(cfg *config.Config) *MemoryRepository {
	return &MemoryRepository{
		MemoryStorage: make(map[string]string),
		cfg:           cfg,
	}
}
