package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword generates a secure bcrypt hash of the provided password.
// Uses a cost factor of 14 for optimal security-performance balance.
//
// Parameters:
//   - password: The plain text password to hash
//
// Returns:
//   - string: The bcrypt hash of the password
//   - error: Any error that occurred during hashing
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash verifies if a plain text password matches its bcrypt hash.
// This is the standard way to validate passwords during authentication.
//
// Parameters:
//   - password: The plain text password to verify
//   - hash: The bcrypt hash to compare against
//
// Returns:
//   - bool: true if password matches the hash, false otherwise
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
