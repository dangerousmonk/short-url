package compress

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var CompressedContentTypes = []string{
	"text/html",
	"application/json",
}

func decompress(encodings string, body io.ReadCloser) (io.ReadCloser, error) {
	encodings = strings.ToLower(encodings)
	if !strings.Contains(encodings, "gzip") {
		return nil, fmt.Errorf("unsupported compression method: %s", encodings)
	}
	reader, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func DecompressMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		encodings := r.Header.Get("Content-Encoding")
		if encodings == "" {
			next.ServeHTTP(w, r)
			return
		}
		reader, err := decompress(encodings, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer reader.Close()
		r.Body = reader
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
