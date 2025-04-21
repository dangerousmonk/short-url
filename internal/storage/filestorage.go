package storage

import (
	"encoding/json"
	"os"
)

type FileWriter struct {
	file    *os.File
	encoder *json.Encoder
}

type FileReader struct {
	file    *os.File
	decoder *json.Decoder
}

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

func (fw *FileWriter) WriteData(data *Row) error {
	return fw.encoder.Encode(&data)
}

func (fw *FileWriter) Close() error {
	return fw.file.Close()
}

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

func (fr *FileReader) ReadData(s *MapStorage) (*Row, error) {
	var row Row
	for fr.decoder.Decode(&row) == nil {
		s.MemoryStorage[row.ShortURL] = row.OriginalURL
	}
	return &row, nil
}

func (fr *FileReader) Close() error {
	return fr.file.Close()
}
