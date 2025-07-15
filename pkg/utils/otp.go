package utils

import (
	"crypto/rand"
	"io"
)

// GenerateOTP generates a random numeric OTP of a given length.
func GenerateOTP(length int) (string, error) {
	b := make([]byte, length)
	n, err := io.ReadAtLeast(rand.Reader, b, length)
	if n != length {
		return "", err
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b), nil
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
