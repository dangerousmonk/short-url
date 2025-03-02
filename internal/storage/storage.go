package storage

import (
	"sync"

	"github.com/dangerousmonk/short-url/internal/helpers"
)

type MapStorage struct {
	URLdata map[string]string
	mutex   sync.RWMutex
}

func newMapStorage() *MapStorage {
	return &MapStorage{
		URLdata: make(map[string]string),
	}
}

var AppStorage = newMapStorage()

func (s *MapStorage) GetFullURL(shortURL string) (FullURL string, isExist bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	FullURL, isExist = s.URLdata[shortURL]
	return FullURL, isExist
}

func (s *MapStorage) AddShortURL(fullURL string) (shortURL string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	hash, err := helpers.HashGenerator()
	if err != nil {
		return
	}
	s.URLdata[hash] = fullURL
	return hash
}
