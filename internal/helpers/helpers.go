package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
)

func HashGenerator() (string, error) {
	bytes := make([]byte, 4)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func IsURLValid(rawURL string) bool {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return parsedURL.Scheme == "https" || parsedURL.Scheme == "http"
}
