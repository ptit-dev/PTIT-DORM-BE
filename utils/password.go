package utils

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

// GenerateStrongPassword returns a random string of n characters, base64 encoded
func GenerateStrongPassword(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "randompass123"
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
