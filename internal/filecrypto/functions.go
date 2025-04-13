package anbuFileCrypto

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/tanq16/anbu/utils"
)

func EncryptFile(inputPath string, password string) error {
	logger := utils.GetLogger("filecrypto")
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	outputPath := inputPath + ".enc"
	err = executeOpenSSLEncrypt(content, outputPath, password)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}
	logger.Debug().Str("input", inputPath).Str("output", outputPath).Msg("file encrypted")
	return nil
}

func DecryptFile(inputPath string, password string) error {
	logger := utils.GetLogger("filecrypto")
	if !strings.HasSuffix(inputPath, ".enc") {
		return fmt.Errorf("input file must have .enc extension")
	}
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	outputPath := strings.TrimSuffix(inputPath, ".enc")
	err = executeOpenSSLDecrypt(content, outputPath, password)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}
	logger.Debug().Str("input", inputPath).Str("output", outputPath).Msg("file decrypted")
	return nil
}

func executeOpenSSLEncrypt(content []byte, outputPath string, password string) error {
	block, err := aes.NewCipher([]byte(getKey(password)))
	if err != nil {
		return err
	}
	paddedContent := pkcs7Pad(content, aes.BlockSize)
	encryptedContent := make([]byte, len(paddedContent))
	for i := 0; i < len(paddedContent); i += aes.BlockSize {
		block.Encrypt(encryptedContent[i:i+aes.BlockSize], paddedContent[i:i+aes.BlockSize])
	}
	encoder := base64.StdEncoding
	encodedLen := encoder.EncodedLen(len(encryptedContent))
	encoded := make([]byte, encodedLen)
	encoder.Encode(encoded, encryptedContent)
	return os.WriteFile(outputPath, encoded, 0644)
}

func executeOpenSSLDecrypt(content []byte, outputPath string, password string) error {
	decoder := base64.StdEncoding
	decodedLen := decoder.DecodedLen(len(content))
	decoded := make([]byte, decodedLen)
	n, err := decoder.Decode(decoded, content)
	if err != nil {
		return err
	}
	decoded = decoded[:n]
	block, err := aes.NewCipher([]byte(getKey(password)))
	if err != nil {
		return err
	}
	decrypted := make([]byte, len(decoded))
	for i := 0; i < len(decoded); i += aes.BlockSize {
		block.Decrypt(decrypted[i:i+aes.BlockSize], decoded[i:i+aes.BlockSize])
	}
	unpadded := pkcs7Unpad(decrypted)
	return os.WriteFile(outputPath, unpadded, 0644)
}

func getKey(password string) []byte {
	key := make([]byte, 32)
	copy(key, []byte(password))
	return key
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

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
