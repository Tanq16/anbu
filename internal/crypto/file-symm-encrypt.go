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

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
)

func EncryptFileSymmetric(inputPath string, password string) {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read input file")
	}
	// Generate a 32-byte key from the password
	hash := sha256.Sum256([]byte(password))
	key := hash[:]
	// Create a new AES cipher block in GCM mode and generate a nonce
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Fatal().Err(err).Msg("failed to generate nonce")
	}
	// Encrypt and encode the data
	ciphertext := gcm.Seal(nonce, nonce, content, nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	// Write encrypted data to output file
	outputPath := inputPath + ".enc"
	os.WriteFile(outputPath, []byte(encoded), 0644)
	fmt.Printf("\nFile encrypted: %s\n", u.FSuccess(outputPath))
}

func DecryptFileSymmetric(inputPath string, password string) {
	encContent, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read input file")
	}
	// Decode from base64 and generate key from password
	decoded, err := base64.StdEncoding.DecodeString(string(encContent))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to decode base64 content")
	}
	hash := sha256.Sum256([]byte(password))
	key := hash[:]
	// Create a new AES cipher block in GCM mode and extract nonce
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()
	if len(decoded) < nonceSize {
		log.Fatal().Msg("ciphertext too short")
	}
	// Extract nonce and ciphertext and decrypt
	nonce, ciphertext := decoded[:nonceSize], decoded[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to decrypt")
	}
	// Write decrypted data to output file
	outputPath := strings.TrimSuffix(inputPath, ".enc")
	os.WriteFile(outputPath, plaintext, 0644)
	fmt.Printf("\nFile decrypted: %s\n", u.FSuccess(outputPath))
}
