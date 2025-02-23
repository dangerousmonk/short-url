package helpers

import (
	"crypto/rand"
	"encoding/hex"
)

func HashGenerator() (string, error) {
	bytes := make([]byte, 4)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	hash := hex.EncodeToString(bytes)
	return hash[:8], nil
}
