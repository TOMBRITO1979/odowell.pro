package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"
)

var encryptionKey []byte

// InitEncryption initializes the encryption key from environment
func InitEncryption() error {
	keyHex := os.Getenv("ENCRYPTION_KEY")
	if keyHex == "" {
		return errors.New("ENCRYPTION_KEY not set")
	}

	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return errors.New("invalid ENCRYPTION_KEY format")
	}

	if len(key) != 32 {
		return errors.New("ENCRYPTION_KEY must be 32 bytes (64 hex chars)")
	}

	encryptionKey = key
	return nil
}

// Encrypt encrypts plaintext using AES-256-GCM
// Returns base64 encoded ciphertext
func Encrypt(plaintext string) (string, error) {
	if len(encryptionKey) == 0 {
		if err := InitEncryption(); err != nil {
			return "", err
		}
	}

	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64 encoded ciphertext using AES-256-GCM
func Decrypt(ciphertext string) (string, error) {
	if len(encryptionKey) == 0 {
		if err := InitEncryption(); err != nil {
			return "", err
		}
	}

	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// IsEncrypted checks if a string appears to be encrypted (base64 with specific length)
func IsEncrypted(s string) bool {
	if s == "" {
		return false
	}
	// Encrypted strings are base64 and at least 28 chars (12 byte nonce + 16 byte tag minimum)
	if len(s) < 28 {
		return false
	}
	// Try to decode as base64
	_, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return false
	}
	// Check if it looks like a Stripe key or plain text (not encrypted)
	if strings.HasPrefix(s, "sk_") || strings.HasPrefix(s, "pk_") ||
	   strings.HasPrefix(s, "rk_") || strings.HasPrefix(s, "whsec_") {
		return false
	}
	return true
}

// EncryptIfNeeded encrypts only if not already encrypted
func EncryptIfNeeded(plaintext string) (string, error) {
	if IsEncrypted(plaintext) {
		return plaintext, nil
	}
	return Encrypt(plaintext)
}

// DecryptIfNeeded decrypts only if encrypted
func DecryptIfNeeded(ciphertext string) (string, error) {
	if !IsEncrypted(ciphertext) {
		return ciphertext, nil
	}
	return Decrypt(ciphertext)
}

// HashAPIKey creates a SHA-256 hash of an API key
// Used for secure storage - the original key cannot be recovered
func HashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// VerifyAPIKey compares a plain API key with its hash using constant-time comparison
// Returns true if they match
func VerifyAPIKey(plainKey, hashedKey string) bool {
	computedHash := HashAPIKey(plainKey)
	return subtle.ConstantTimeCompare([]byte(computedHash), []byte(hashedKey)) == 1
}

// IsAPIKeyHash checks if a string looks like an API key hash (64 hex chars = SHA-256)
func IsAPIKeyHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// EncryptAES encrypts data using AES-256-GCM with a custom key
func EncryptAES(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptAES decrypts data using AES-256-GCM with a custom key
func DecryptAES(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertextBytes, nil)
}
