package anbuCrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type SecretsStore struct {
	Secrets map[string]string `json:"secrets"`
}

func InitializeSecretsStore(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		store := &SecretsStore{
			Secrets: make(map[string]string),
		}
		if err := saveSecretsStore(store, filePath); err != nil {
			return fmt.Errorf("failed to create secrets store: %w", err)
		}
	}
	return nil
}

func saveSecretsStore(store *SecretsStore, filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal store: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write secrets file: %w", err)
	}
	return nil
}

func ListSecrets(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}
	var store SecretsStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("failed to parse secrets file: %w", err)
	}
	var secrets []string
	for id := range store.Secrets {
		secrets = append(secrets, id)
	}
	return secrets, nil
}

func GetSecret(filePath, secretID, password string) (string, error) {
	store, err := loadSecretsStore(filePath)
	if err != nil {
		return "", err
	}
	encryptedValue, exists := store.Secrets[secretID]
	if !exists {
		return "", fmt.Errorf("secret '%s' not found", secretID)
	}
	return decryptString(encryptedValue, password)
}

func SetSecret(filePath, secretID, value, password string) error {
	store, err := loadSecretsStore(filePath)
	if err != nil {
		return err
	}
	encryptedValue, err := encryptString(value, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}
	store.Secrets[secretID] = encryptedValue
	return saveSecretsStore(store, filePath)
}

func DeleteSecret(filePath, secretID string) error {
	store, err := loadSecretsStore(filePath)
	if err != nil {
		return err
	}
	if _, exists := store.Secrets[secretID]; !exists {
		return fmt.Errorf("secret '%s' not found", secretID)
	}
	delete(store.Secrets, secretID)
	return saveSecretsStore(store, filePath)
}

func ImportSecrets(filePath, importFilePath string, password string) error {
	importData, err := os.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}
	var importStore SecretsStore
	if err := json.Unmarshal(importData, &importStore); err != nil {
		return fmt.Errorf("failed to parse import file: %w", err)
	}
	currentStore, err := loadSecretsStore(filePath)
	if err != nil {
		return err
	}
	for secretID, secretValue := range importStore.Secrets {
		if err := setSecret(currentStore, filePath, secretID, secretValue, password); err != nil {
			return fmt.Errorf("failed to encrypt and save secret '%s': %w", secretID, err)
		}
	}
	return nil
}

func ExportSecrets(filePath, exportFilePath string, password string) error {
	store, err := loadSecretsStore(filePath)
	if err != nil {
		return err
	}
	exportStore := &SecretsStore{
		Secrets: make(map[string]string),
	}
	for id, encryptedValue := range store.Secrets {
		decryptedValue, err := decryptString(encryptedValue, password)
		if err != nil {
			return fmt.Errorf("failed to decrypt secret '%s': %w", id, err)
		}
		exportStore.Secrets[id] = decryptedValue
	}
	exportData, err := json.MarshalIndent(exportStore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal export data: %w", err)
	}
	return os.WriteFile(exportFilePath, exportData, 0600)
}

// Private helper methods for encryption/decryption
func loadSecretsStore(filePath string) (*SecretsStore, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}
	var store SecretsStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("failed to parse secrets file: %w", err)
	}
	return &store, nil
}

func encryptString(value, password string) (string, error) {
	key := generateKeyFromPassword(password)
	encryptedBytes, err := encryptData([]byte(value), key)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt string: %w", err)
	}
	encodedValue := base64.StdEncoding.EncodeToString(encryptedBytes)
	return encodedValue, nil
}

func decryptString(encryptedValue, password string) (string, error) {
	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decode string: %w", err)
	}
	key := generateKeyFromPassword(password)
	decryptedBytes, err := decryptData(encryptedBytes, key)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt string: %w", err)
	}
	return string(decryptedBytes), nil
}

func generateKeyFromPassword(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

func encryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func decryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func setSecret(store *SecretsStore, filePath, secretID, value, password string) error {
	encryptedValue, err := encryptString(value, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}
	store.Secrets[secretID] = encryptedValue
	return saveSecretsStore(store, filePath)
}
