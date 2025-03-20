package storage

import (
	"strconv"
	"sync"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/helpers"
)

type Row struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type MapStorage struct {
	URLdata map[string]string
	mutex   sync.RWMutex
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		URLdata: make(map[string]string),
	}
}

func (s *MapStorage) GetFullURL(shortURL string) (FullURL string, isExist bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	FullURL, isExist = s.URLdata[shortURL]
	return FullURL, isExist
}

func (s *MapStorage) AddShortURL(fullURL string, storagePath string) (shortURL string, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	shortURL, err = helpers.HashGenerator()
	if err != nil {
		return
	}
	s.URLdata[shortURL] = fullURL
	urlData := Row{UUID: strconv.Itoa(len(s.URLdata)), ShortURL: shortURL, OriginalURL: fullURL}

	writer, err := NewWriter(storagePath)
	if err != nil {
		return
	}
	defer writer.Close()

	if err = writer.WriteData(&urlData); err != nil {
		return
	}
	return shortURL, nil
}

func (s *MapStorage) LoadFromFile(cfg *config.Config) error {
	reader, err := NewFileReader(cfg.StorageFilePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = reader.ReadData(s)
	if err != nil {
		return err
	}
	return nil
}
