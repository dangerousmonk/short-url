package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/dangerousmonk/short-url/internal/logging"
)

var CompressedContentTypes = []string{
	"text/html",
	"application/json",
}

func decompress(body io.ReadCloser) (io.ReadCloser, error) {
	reader, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

// Decompress Middleware is the middleware function that adds gzip-compression functionality
func DecompressMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		encodings := r.Header.Get("Content-Encoding")
		if encodings == "" {
			next.ServeHTTP(w, r)
			return
		}

		encodings = strings.ToLower(encodings)
		if !strings.Contains(encodings, "gzip") {
			http.Error(w, "Unsupported compression method", http.StatusUnsupportedMediaType)
			return
		}

		reader, err := decompress(r.Body)
		if err != nil {
			logging.Log.Warnf("Failed to decompress | err=%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer reader.Close()

		// Close old request.Body before substitution
		if r.Body != nil {
			if err := r.Body.Close(); err != nil {
				logging.Log.Warnf("Failed to close original body | err=%v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		r.Body = reader
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
