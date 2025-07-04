package memory

import (
	"encoding/json"
	"os"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/models"
)

// FileWriter describes struct to write URL data to file
type FileWriter struct {
	file    *os.File
	encoder *json.Encoder
}

// NewWriter is a helper function that tries to open file for storing URL data
func NewWriter(fileName string) (*FileWriter, error) {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileWriter{
		file:    f,
		encoder: json.NewEncoder(f),
	}, nil
}

// WriteData function of FileWriter is used to write to stream.
func (fw *FileWriter) WriteData(data *models.URLData) error {
	return fw.encoder.Encode(&data)
}

// Close function of FileWriter is used to close file after we done writing to it.
func (fw *FileWriter) Close() error {
	return fw.file.Close()
}

// FileReader describes struct to read URL data from file.
type FileReader struct {
	file    *os.File
	decoder *json.Decoder
}

// NewFileReader is a helper function that tries to open file in a read mode if it exists.
func NewFileReader(fileName string) (*FileReader, error) {
	f, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &FileReader{
		file:    f,
		decoder: json.NewDecoder(f),
	}, nil
}

// ReadData function of FileReader is used to read URL data from file into memory storage.
func (fr *FileReader) ReadData(r *MemoryRepository) (*models.URLData, error) {
	var row models.URLData
	for fr.decoder.Decode(&row) == nil {
		r.MemoryStorage[row.ShortURL] = row.OriginalURL
	}
	return &row, nil
}

// Close function of FileReader is used to close file after we done reading from it.
func (fr *FileReader) Close() error {
	return fr.file.Close()
}

// LoadFromFile is used to load data into memory from file.
func LoadFromFile(r *MemoryRepository, cfg *config.Config) error {
	reader, err := NewFileReader(cfg.StorageFilePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, err = reader.ReadData(r)
	if err != nil {
		return err
	}
	return nil
}
