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

	"github.com/tanq16/anbu/utils"
)

func EncryptSymmetric(inputPath string, password string) error {
	logger := utils.GetLogger("filecrypto")
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Generate a 32-byte key from the password
	key := deriveKey(password)

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create a GCM cipher mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Generate a nonce (IV)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, content, nil)

	// Encode the ciphertext in base64 for storage
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	// Write encrypted data to output file
	outputPath := inputPath + ".enc"
	err = os.WriteFile(outputPath, []byte(encoded), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	logger.Debug().Str("input", inputPath).Str("output", outputPath).Msg("file encrypted with AES-GCM")
	return nil
}

func DecryptSymmetric(inputPath string, password string) error {
	logger := utils.GetLogger("filecrypto")
	if !strings.HasSuffix(inputPath, ".enc") {
		return fmt.Errorf("input file must have .enc extension")
	}

	// Read the encrypted content
	encContent, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(string(encContent))
	if err != nil {
		return fmt.Errorf("failed to decode base64 content: %w", err)
	}

	// Generate key from password
	key := deriveKey(password)

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create a GCM cipher mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Extract nonce size
	nonceSize := gcm.NonceSize()
	if len(decoded) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := decoded[:nonceSize], decoded[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	// Write decrypted data to output file
	outputPath := strings.TrimSuffix(inputPath, ".enc")
	err = os.WriteFile(outputPath, plaintext, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	logger.Debug().Str("input", inputPath).Str("output", outputPath).Msg("file decrypted with AES-GCM")
	return nil
}

func deriveKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}
