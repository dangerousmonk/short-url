package helpers

import (
	"crypto/rand"
	"encoding/hex"
)

// Pre-allocate a reusable buffer to avoid allocations
var hashBuf = make([]byte, 4)

// HashGenerator is helper funtion to generate random hash
func HashGenerator() (string, error) {
	if _, err := rand.Read(hashBuf); err != nil {
		return "", err
	}
	return hex.EncodeToString(hashBuf), nil
}

// IsURLValid is helper funtion that checks if provided string is vaild URL
func IsURLValid(s string) bool {
	if len(s) < 7 {
		return false
	}

	// Check scheme prefix without full parsing
	switch {
	case s[0] == 'h' && s[1] == 't' && s[2] == 't' && s[3] == 'p':

		if s[4] == ':' && s[5] == '/' && s[6] == '/' {
			return true
		}
		if s[4] == 's' && s[5] == ':' && s[6] == '/' && s[7] == '/' {
			return true
		}
	}
	return false
}
