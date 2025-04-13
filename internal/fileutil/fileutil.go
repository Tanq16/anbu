package fileutil

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/tanq16/anbu/utils"
)

// EncryptFile encrypts a file using AES-256-ECB
func EncryptFile(inputPath string, password string) error {
	logger := utils.GetLogger("fileutil")

	// Read input file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Create output path
	outputPath := inputPath + ".enc"

	// Execute OpenSSL command
	err = executeOpenSSLEncrypt(content, outputPath, password)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	logger.Info().Str("input", inputPath).Str("output", outputPath).Msg("file encrypted successfully")
	return nil
}

// DecryptFile decrypts a file that was encrypted using AES-256-ECB
func DecryptFile(inputPath string, password string) error {
	logger := utils.GetLogger("fileutil")

	// Check if file has .enc extension
	if !strings.HasSuffix(inputPath, ".enc") {
		return fmt.Errorf("input file must have .enc extension")
	}

	// Read input file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Create output path - remove .enc extension
	outputPath := strings.TrimSuffix(inputPath, ".enc")

	// Execute OpenSSL command
	err = executeOpenSSLDecrypt(content, outputPath, password)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	logger.Info().Str("input", inputPath).Str("output", outputPath).Msg("file decrypted successfully")
	return nil
}

// executeOpenSSLEncrypt implements AES-256-ECB encryption with base64 encoding
func executeOpenSSLEncrypt(content []byte, outputPath string, password string) error {
	// Implement OpenSSL-compatible AES-256-ECB encryption
	// For production use, consider using a proper Go crypto package instead
	block, err := aes.NewCipher([]byte(getKey(password)))
	if err != nil {
		return err
	}

	// Pad the content to a multiple of the block size
	paddedContent := pkcs7Pad(content, aes.BlockSize)

	// Encrypt the padded content
	encryptedContent := make([]byte, len(paddedContent))

	// ECB mode (note: ECB is not secure for most uses, but matching the shell script)
	for i := 0; i < len(paddedContent); i += aes.BlockSize {
		block.Encrypt(encryptedContent[i:i+aes.BlockSize], paddedContent[i:i+aes.BlockSize])
	}

	// Base64 encode the encrypted content
	encoder := base64.StdEncoding
	encodedLen := encoder.EncodedLen(len(encryptedContent))
	encoded := make([]byte, encodedLen)
	encoder.Encode(encoded, encryptedContent)

	// Write to output file
	return os.WriteFile(outputPath, encoded, 0644)
}

// executeOpenSSLDecrypt implements AES-256-ECB decryption with base64 decoding
func executeOpenSSLDecrypt(content []byte, outputPath string, password string) error {
	// Base64 decode the content
	decoder := base64.StdEncoding
	decodedLen := decoder.DecodedLen(len(content))
	decoded := make([]byte, decodedLen)
	n, err := decoder.Decode(decoded, content)
	if err != nil {
		return err
	}
	decoded = decoded[:n]

	// Create a new AES cipher
	block, err := aes.NewCipher([]byte(getKey(password)))
	if err != nil {
		return err
	}

	// Decrypt the content
	decrypted := make([]byte, len(decoded))

	// ECB mode
	for i := 0; i < len(decoded); i += aes.BlockSize {
		block.Decrypt(decrypted[i:i+aes.BlockSize], decoded[i:i+aes.BlockSize])
	}

	// Unpad the content
	unpadded := pkcs7Unpad(decrypted)

	// Write to output file
	return os.WriteFile(outputPath, unpadded, 0644)
}

// getKey creates a 32-byte key from the password
func getKey(password string) []byte {
	key := make([]byte, 32) // AES-256 requires a 32-byte key
	copy(key, []byte(password))
	return key
}

// pkcs7Pad adds PKCS#7 padding to a block
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// pkcs7Unpad removes PKCS#7 padding from a block
func pkcs7Unpad(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return nil
	}
	padding := int(data[length-1])
	if padding > length {
		return data // Invalid padding
	}
	return data[:length-padding]
}

// ProcessLinearOperation executes a command for each number in the range
func ProcessLinearOperation(count int, command string) error {
	logger := utils.GetLogger("fileutil")

	if count < 0 {
		return fmt.Errorf("count must be non-negative")
	}

	for i := 0; i <= count; i++ {
		// Replace $i with the current value in the command
		cmdToRun := strings.ReplaceAll(command, "$i", strconv.Itoa(i))

		logger.Debug().Str("command", cmdToRun).Msg("executing command")

		// Execute the command using the shell
		// Note: This is potentially dangerous and should be used with caution
		// For a production tool, you might want to limit the commands that can be executed
		// or implement a safer way to achieve the same functionality
		cmd := exec.Command("sh", "-c", cmdToRun)
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf

		err := cmd.Run()
		if err != nil {
			logger.Error().Err(err).Str("stderr", stderrBuf.String()).Msg("command execution failed")
			return fmt.Errorf("command execution failed: %w", err)
		}

		if stdoutBuf.Len() > 0 {
			logger.Info().Str("output", stdoutBuf.String()).Msg("command output")
		}
	}

	return nil
}
