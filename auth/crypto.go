package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"strings"

	"golang.org/x/crypto/argon2"
)

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func ComputeLookupHash(email string, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(email))
	return h.Sum(nil)
}

func GeneratePasswordSalt() []byte {
	salt := make([]byte, 32)
	rand.Read(salt)
	return salt
}

func HashPassword(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
}

func VerifyPasswordHash(password, hash, salt []byte) bool {
	derived := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(derived, hash) == 1
}

func GenerateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return bytes
}

// EncryptedData represents encrypted data with IV and authentication tag
type EncryptedData struct {
	Ciphertext []byte
	IV         []byte
	Tag        []byte
}

// EncryptEmail encrypts an email using AES-GCM
func EncryptEmail(email string, key []byte) (*EncryptedData, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate random IV
	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	// Encrypt the email
	ciphertext := gcm.Seal(nil, iv, []byte(email), nil)

	// Split ciphertext and tag (last 16 bytes)
	tagSize := gcm.Overhead()
	if len(ciphertext) < tagSize {
		return nil, err
	}

	data := ciphertext[:len(ciphertext)-tagSize]
	tag := ciphertext[len(ciphertext)-tagSize:]

	return &EncryptedData{
		Ciphertext: data,
		IV:         iv,
		Tag:        tag,
	}, nil
}

// DecryptEmail decrypts an encrypted email
func DecryptEmail(encrypted *EncryptedData, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	fullCiphertext := append(encrypted.Ciphertext, encrypted.Tag...)

	plaintext, err := gcm.Open(nil, encrypted.IV, fullCiphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GenerateEncryptionKey generates a 32-byte AES-256 encryption key
func GenerateEncryptionKey() []byte {
	return GenerateRandomBytes(32)
}

// GeneratePIN generates a 6-character alphanumeric PIN (lowercase letters and digits)
func GeneratePIN() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const pinLength = 6

	pin := make([]byte, pinLength)
	randomBytes := make([]byte, pinLength)

	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to a simpler approach if crypto/rand fails
		panic("failed to generate random bytes for PIN")
	}

	for i := 0; i < pinLength; i++ {
		pin[i] = charset[int(randomBytes[i])%len(charset)]
	}

	return string(pin)
}

// NormalizePIN normalizes a PIN to lowercase for case-insensitive comparison
// PINs are always stored and compared in lowercase
func NormalizePIN(pin string) string {
	return strings.ToLower(strings.TrimSpace(pin))
}

// ComputePINLookupHash computes a lookup hash for PIN (similar to email lookup)
// The PIN should be normalized before hashing
func ComputePINLookupHash(pin string, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(pin))
	return h.Sum(nil)
}

// GenerateSecurePassword generates a cryptographically secure random password
func GenerateSecurePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%&*()_+-="
	password := make([]byte, length)
	randomBytes := make([]byte, length)

	if _, err := rand.Read(randomBytes); err != nil {
		panic("failed to generate random bytes for password")
	}

	for i := 0; i < length; i++ {
		password[i] = charset[int(randomBytes[i])%len(charset)]
	}

	return string(password)
}
