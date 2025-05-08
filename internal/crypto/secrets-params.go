package anbuCrypto

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	u "github.com/tanq16/anbu/utils"
	"golang.org/x/term"
)

type SecretsStore struct {
	Secrets    map[string]string `json:"secrets"`
	Parameters map[string]string `json:"parameters"`
}

func InitializeSecretsStore(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		store := &SecretsStore{
			Secrets:    make(map[string]string),
			Parameters: make(map[string]string),
		}
		if err := saveSecretsStore(store, filePath); err != nil {
			return fmt.Errorf("failed to create secrets store: %w", err)
		}
	}
	return nil
}

func loadSecretsStoreStructure(filePath string) (*SecretsStore, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}
	store := &SecretsStore{
		Secrets:    make(map[string]string),
		Parameters: make(map[string]string),
	}
	if err := json.Unmarshal(data, store); err != nil {
		return nil, fmt.Errorf("failed to parse secrets file: %w", err)
	}
	return store, nil
}

func loadAndDecryptSecret(filePath, secretID string, password string) (string, error) {
	store, err := loadSecretsStoreStructure(filePath)
	if err != nil {
		return "", err
	}
	encryptedValue, exists := store.Secrets[secretID]
	if !exists {
		return "", fmt.Errorf("secret '%s' not found", secretID)
	}
	decryptedBytes, err := decryptString(encryptedValue, password)
	if err != nil {
		return "", err
	}
	return decryptedBytes, nil
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

func encryptAndSaveSecret(filePath, secretID, value, password string) error {
	store, err := loadSecretsStoreStructure(filePath)
	if err != nil {
		return err
	}
	// Encrypt the value
	encryptedValue, err := encryptString(value, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}
	// Update the store
	store.Secrets[secretID] = encryptedValue
	// Save the store
	if err := saveSecretsStore(store, filePath); err != nil {
		return err
	}
	return nil
}

func encryptString(value, password string) (string, error) {
	key := generateKeyFromPassword(password)
	// Encrypt the value
	encryptedBytes, err := encryptData([]byte(value), key)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt string: %w", err)
	}
	// Encode to base64
	encodedValue := base64.StdEncoding.EncodeToString(encryptedBytes)
	return encodedValue, nil
}

func decryptString(encryptedValue, password string) (string, error) {
	// Decode from base64
	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decode string: %w", err)
	}
	// Generate key from password
	key := generateKeyFromPassword(password)
	// Decrypt the value
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
	// Encrypt and prepend nonce
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

func getPassword() (string, error) {
	password := os.Getenv("ANBUPW")
	if password != "" {
		return password, nil
	}
	fmt.Print("Enter password for secrets: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add a newline after input
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return string(passwordBytes), nil
}

func ListSecrets(filePath string) error {
	store, err := loadSecretsStoreStructure(filePath)
	if err != nil {
		return err
	}
	table := u.NewTable([]string{"Secrets", "Parameters"})
	maxRows := max(len(store.Secrets), len(store.Parameters))
	secretKeys := make([]string, 0, len(store.Secrets))
	for id := range store.Secrets {
		secretKeys = append(secretKeys, id)
	}
	paramKeys := make([]string, 0, len(store.Parameters))
	for id := range store.Parameters {
		paramKeys = append(paramKeys, id)
	}
	for i := range maxRows {
		var secretID, paramID string
		if i < len(secretKeys) {
			secretID = secretKeys[i]
		}
		if i < len(paramKeys) {
			paramID = paramKeys[i]
		}
		table.Rows = append(table.Rows, []string{secretID, paramID})
	}
	fmt.Println()
	table.PrintTable(false)
	fmt.Printf("\nTotal: %d secrets, %d parameters\n", len(store.Secrets), len(store.Parameters))
	return nil
}

func GetSecret(filePath, secretID string) error {
	password, err := getPassword()
	if err != nil {
		return err
	}
	value, err := loadAndDecryptSecret(filePath, secretID, password)
	if err != nil {
		return err
	}
	fmt.Println(value)
	return nil
}

func GetParameter(filePath, paramID string) error {
	store, err := loadSecretsStoreStructure(filePath)
	if err != nil {
		return err
	}
	value, exists := store.Parameters[paramID]
	if !exists {
		return fmt.Errorf("parameter '%s' not found", paramID)
	}
	fmt.Println(value)
	return nil
}

func SetSecret(filePath, secretID string) error {
	fmt.Printf("Enter value for secret '%s': ", secretID)
	value, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add a newline after input
	if err != nil {
		return fmt.Errorf("failed to read secret value: %w", err)
	}
	password, err := getPassword()
	if err != nil {
		return err
	}
	if err := encryptAndSaveSecret(filePath, secretID, string(value), password); err != nil {
		return err
	}
	u.PrintSuccess(fmt.Sprintf("Secret '%s' set successfully", secretID))
	return nil
}

func SetParameter(filePath, paramID string) error {
	store, err := loadSecretsStoreStructure(filePath)
	if err != nil {
		return err
	}
	fmt.Printf("Enter value for parameter '%s': ", paramID)
	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read parameter value: %w", err)
	}
	value = strings.TrimSpace(value)
	store.Parameters[paramID] = value
	if err := saveSecretsStore(store, filePath); err != nil {
		return err
	}
	u.PrintSuccess(fmt.Sprintf("Parameter '%s' set successfully", paramID))
	return nil
}

func DeleteSecret(filePath, secretID string) error {
	store, err := loadSecretsStoreStructure(filePath)
	if err != nil {
		return err
	}
	if _, exists := store.Secrets[secretID]; !exists {
		return fmt.Errorf("secret '%s' not found", secretID)
	}
	delete(store.Secrets, secretID)
	if err := saveSecretsStore(store, filePath); err != nil {
		return err
	}
	u.PrintSuccess(fmt.Sprintf("Secret '%s' deleted successfully", secretID))
	return nil
}

func DeleteParameter(filePath, paramID string) error {
	store, err := loadSecretsStoreStructure(filePath)
	if err != nil {
		return err
	}
	if _, exists := store.Parameters[paramID]; !exists {
		return fmt.Errorf("parameter '%s' not found", paramID)
	}
	delete(store.Parameters, paramID)
	if err := saveSecretsStore(store, filePath); err != nil {
		return err
	}
	u.PrintSuccess(fmt.Sprintf("Parameter '%s' deleted successfully", paramID))
	return nil
}

func ImportSecrets(filePath, importFilePath string) error {
	importData, err := os.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}
	var importStore SecretsStore
	if err := json.Unmarshal(importData, &importStore); err != nil {
		return fmt.Errorf("failed to parse import file: %w", err)
	}
	currentStore, err := loadSecretsStoreStructure(filePath)
	if err != nil {
		return err
	}
	// Copy parameters from import to current store
	if currentStore.Parameters == nil {
		currentStore.Parameters = make(map[string]string)
	}
	maps.Copy(currentStore.Parameters, importStore.Parameters)
	if err := saveSecretsStore(currentStore, filePath); err != nil {
		return err
	}
	// Process secrets
	password, err := getPassword()
	if err != nil {
		return err
	}
	for secretID, secretValue := range importStore.Secrets {
		if err := encryptAndSaveSecret(filePath, secretID, secretValue, password); err != nil {
			return fmt.Errorf("failed to encrypt and save secret '%s': %w", secretID, err)
		}
	}
	u.PrintSuccess(fmt.Sprintf("Imported %d secrets and %d parameters successfully", len(importStore.Secrets), len(importStore.Parameters)))
	return nil
}

func ExportSecrets(filePath, exportFilePath string) error {
	store, err := loadSecretsStoreStructure(filePath)
	if err != nil {
		return err
	}
	exportStore := &SecretsStore{
		Secrets:    make(map[string]string),
		Parameters: store.Parameters,
	}
	password, err := getPassword()
	if err != nil {
		return err
	}
	for id := range store.Secrets {
		decryptedValue, err := loadAndDecryptSecret(filePath, id, password)
		if err != nil {
			return fmt.Errorf("failed to decrypt secret '%s': %w", id, err)
		}
		exportStore.Secrets[id] = decryptedValue
	}
	exportData, err := json.MarshalIndent(exportStore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal export data: %w", err)
	}
	if err := os.WriteFile(exportFilePath, exportData, 0600); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}
	u.PrintSuccess(fmt.Sprintf("Exported %d secrets and %d parameters to %s", len(exportStore.Secrets), len(exportStore.Parameters), exportFilePath))
	return nil
}
