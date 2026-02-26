package anbuCrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

var encryptionSalt = []byte("anbu-file-crypt-v1")

func EncryptFileSymmetric(inputPath string, password string) (string, error) {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read input file: %w", err)
	}
	key := pbkdf2.Key([]byte(password), encryptionSalt, 100000, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, content, nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	outputPath := inputPath + ".enc"
	if err := os.WriteFile(outputPath, []byte(encoded), 0644); err != nil {
		return "", fmt.Errorf("failed to write encrypted file: %w", err)
	}
	return outputPath, nil
}

func DecryptFileSymmetric(inputPath string, password string) (string, error) {
	encContent, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read input file: %w", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(string(encContent))
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 content: %w", err)
	}
	key := pbkdf2.Key([]byte(password), encryptionSalt, 100000, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(decoded) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := decoded[:nonceSize], decoded[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	outputPath := strings.TrimSuffix(inputPath, ".enc")
	if err := os.WriteFile(outputPath, plaintext, 0644); err != nil {
		return "", fmt.Errorf("failed to write decrypted file: %w", err)
	}
	return outputPath, nil
}
