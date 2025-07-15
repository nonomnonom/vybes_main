package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// Encrypt encrypts plaintext data using AES-GCM encryption with the provided key.
// The encrypted data is returned as a base64-encoded string for safe storage/transmission.
// AES-GCM provides both confidentiality and authenticity.
//
// Parameters:
//   - text: The plaintext string to encrypt
//   - key: The encryption key (should be 32 bytes for AES-256)
//
// Returns:
//   - string: Base64-encoded encrypted data
//   - error: Any error that occurred during encryption
func Encrypt(text string, key []byte) (string, error) {
	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// Encrypt and seal
	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded encrypted data using AES-GCM decryption.
// The function extracts the nonce from the beginning of the ciphertext and
// verifies the authenticity of the encrypted data.
//
// Parameters:
//   - encryptedText: Base64-encoded encrypted data
//   - key: The decryption key (must match the encryption key)
//
// Returns:
//   - string: The decrypted plaintext
//   - error: Any error that occurred during decryption
func Decrypt(encryptedText string, key []byte) (string, error) {
	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Check ciphertext length
	if len(ciphertext) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce and decrypt
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GenerateRandomString creates a cryptographically secure random string
// of the specified length using characters from the provided charset.
// Uses crypto/rand for secure random number generation.
//
// Parameters:
//   - length: The desired length of the random string
//   - charset: The character set to use for generation (e.g., "abcdefghijklmnopqrstuvwxyz")
//
// Returns:
//   - string: A random string of the specified length
//   - error: Any error that occurred during generation
func GenerateRandomString(length int, charset string) (string, error) {
	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to string using charset
	result := make([]byte, length)
	for i := range bytes {
		result[i] = charset[bytes[i]%byte(len(charset))]
	}

	return string(result), nil
}
