package anbuCrypto

// import (
// 	"crypto/aes"
// 	"crypto/cipher"
// 	"crypto/rand"
// 	"crypto/sha256"
// 	"encoding/base64"
// 	"fmt"
// 	"io"
// 	"os"
// 	"strings"
// )

// func EncryptFileSymmetric(inputPath string, password string) error {
// 	content, err := os.ReadFile(inputPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to read input file: %w", err)
// 	}

// 	// Generate a 32-byte key from the password
// 	hash := sha256.Sum256([]byte(password))
// 	key := hash[:]

// 	// Create a new AES cipher block in GCM mode and generate a nonce
// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		return fmt.Errorf("failed to create AES cipher: %w", err)
// 	}
// 	gcm, err := cipher.NewGCM(block)
// 	if err != nil {
// 		return fmt.Errorf("failed to create GCM mode: %w", err)
// 	}
// 	nonce := make([]byte, gcm.NonceSize())
// 	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
// 		return fmt.Errorf("failed to generate nonce: %w", err)
// 	}

// 	// Encrypt and encode the data
// 	ciphertext := gcm.Seal(nonce, nonce, content, nil)
// 	encoded := base64.StdEncoding.EncodeToString(ciphertext)

// 	// Write encrypted data to output file
// 	outputPath := inputPath + ".enc"
// 	err = os.WriteFile(outputPath, []byte(encoded), 0644)
// 	if err != nil {
// 		return fmt.Errorf("failed to write output file: %w", err)
// 	}
// 	logger.Debug().Str("input", inputPath).Str("output", outputPath).Msg("file encrypted")
// 	return nil
// }

// func DecryptFileSymmetric(inputPath string, password string) error {
// 	// if !strings.HasSuffix(inputPath, ".enc") {
// 	// 	return fmt.Errorf("input file must have .enc extension")
// 	// }
// 	encContent, err := os.ReadFile(inputPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to read input file: %w", err)
// 	}

// 	// Decode from base64 and generate key from password
// 	decoded, err := base64.StdEncoding.DecodeString(string(encContent))
// 	if err != nil {
// 		return fmt.Errorf("failed to decode base64 content: %w", err)
// 	}
// 	hash := sha256.Sum256([]byte(password))
// 	key := hash[:]

// 	// Create a new AES cipher block in GCM mode and extract nonce
// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		return fmt.Errorf("failed to create AES cipher: %w", err)
// 	}
// 	gcm, err := cipher.NewGCM(block)
// 	if err != nil {
// 		return fmt.Errorf("failed to create GCM mode: %w", err)
// 	}
// 	nonceSize := gcm.NonceSize()
// 	if len(decoded) < nonceSize {
// 		return fmt.Errorf("ciphertext too short")
// 	}

// 	// Extract nonce and ciphertext and decrypt
// 	nonce, ciphertext := decoded[:nonceSize], decoded[nonceSize:]
// 	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
// 	if err != nil {
// 		return fmt.Errorf("failed to decrypt: %w", err)
// 	}

// 	// Write decrypted data to output file
// 	outputPath := strings.TrimSuffix(inputPath, ".enc")
// 	err = os.WriteFile(outputPath, plaintext, 0644)
// 	if err != nil {
// 		return fmt.Errorf("failed to write output file: %w", err)
// 	}
// 	logger.Debug().Str("input", inputPath).Str("output", outputPath).Msg("file decrypted")
// 	return nil
// }
