// Package memory describes in-memory Repository as well as provides NewMemoryRepository
// function to initialize MemoryRepository
package memory

import (
	"sync"

	"github.com/dangerousmonk/short-url/cmd/config"
)

// MemoryRepository represents in-memory storage.
type MemoryRepository struct {
	cfg           *config.Config
	MemoryStorage map[string]string
	mutex         sync.RWMutex
}

// NewMemoryRepository is a helper function to initalize new in-memory repository.
func NewMemoryRepository(cfg *config.Config) *MemoryRepository {
	return &MemoryRepository{
		MemoryStorage: make(map[string]string),
		cfg:           cfg,
	}
}
