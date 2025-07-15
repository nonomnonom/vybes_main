package utils

import (
	"crypto/rand"
	"fmt"
)

// GenerateOTP creates a cryptographically secure random numeric OTP (One-Time Password)
// of the specified length. Uses crypto/rand for secure random number generation.
//
// Parameters:
//   - length: The number of digits in the OTP (typically 4-8 digits)
//
// Returns:
//   - string: A random numeric OTP of the specified length
//   - error: Any error that occurred during generation
func GenerateOTP(length int) (string, error) {
	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to numeric string
	result := ""
	for _, b := range bytes {
		result += fmt.Sprintf("%d", b%10)
	}

	return result, nil
}
