package auth

import (
	"crypto/rand"
	"encoding/hex"
)

// MakeRefreshToken generates a random 32 bytes hex-encoded string.
func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	encoded := hex.EncodeToString(b)
	return encoded, nil
}
