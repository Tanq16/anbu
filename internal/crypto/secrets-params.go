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
	"os"
	"path/filepath"
	"strings"
	"syscall"

	u "github.com/tanq16/anbu/utils"
	"golang.org/x/term"
)

// SecretsStore represents the structure of the secrets storage
type SecretsStore struct {
	Secrets    map[string]string `json:"secrets"`
	Parameters map[string]string `json:"parameters"`
}

const defaultSecretsFilePath = ".anbu-secrets.json"

// InitializeSecretsStore creates a new empty secrets store if it doesn't exist
func InitializeSecretsStore(filePath string) (*SecretsStore, error) {
	if filePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		filePath = filepath.Join(homeDir, defaultSecretsFilePath)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create a new empty store
		store := &SecretsStore{
			Secrets:    make(map[string]string),
			Parameters: make(map[string]string),
		}

		// Save the empty store
		if err := saveSecretsStore(store, filePath, ""); err != nil {
			return nil, fmt.Errorf("failed to create secrets store: %w", err)
		}
		return store, nil
	}

	// Load existing store
	password, err := getPassword()
	if err != nil {
		return nil, err
	}

	store, err := loadSecretsStore(filePath, password)
	if err != nil {
		return nil, fmt.Errorf("failed to load secrets store: %w", err)
	}

	return store, nil
}

// loadSecretsStore loads and decrypts the secrets store from the specified file
func loadSecretsStore(filePath string, password string) (*SecretsStore, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}

	var encryptedStore struct {
		EncryptedSecrets string            `json:"encrypted_secrets"`
		Parameters       map[string]string `json:"parameters"`
	}

	if err := json.Unmarshal(data, &encryptedStore); err != nil {
		return nil, fmt.Errorf("failed to parse secrets file: %w", err)
	}

	store := &SecretsStore{
		Secrets:    make(map[string]string),
		Parameters: encryptedStore.Parameters,
	}

	// If there are encrypted secrets, decrypt them
	if encryptedStore.EncryptedSecrets != "" {
		decodedSecrets, err := base64.StdEncoding.DecodeString(encryptedStore.EncryptedSecrets)
		if err != nil {
			return nil, fmt.Errorf("failed to decode secrets: %w", err)
		}

		// Generate a 32-byte key from the password
		key := generateKeyFromPassword(password)

		// Decrypt the secrets
		secretsJSON, err := decryptData(decodedSecrets, key)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt secrets: %w", err)
		}

		if err := json.Unmarshal(secretsJSON, &store.Secrets); err != nil {
			return nil, fmt.Errorf("failed to parse decrypted secrets: %w", err)
		}
	}

	return store, nil
}

// saveSecretsStore encrypts and saves the secrets store to the specified file
func saveSecretsStore(store *SecretsStore, filePath string, password string) error {
	// If password is empty, prompt for it
	if password == "" {
		var err error
		password, err = getPassword()
		if err != nil {
			return err
		}
	}

	// Encrypt the secrets
	secretsJSON, err := json.Marshal(store.Secrets)
	if err != nil {
		return fmt.Errorf("failed to marshal secrets: %w", err)
	}

	// Generate a 32-byte key from the password
	key := generateKeyFromPassword(password)

	encryptedSecrets, err := encryptData(secretsJSON, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt secrets: %w", err)
	}

	// Create the encrypted store structure
	encryptedStore := struct {
		EncryptedSecrets string            `json:"encrypted_secrets"`
		Parameters       map[string]string `json:"parameters"`
	}{
		EncryptedSecrets: base64.StdEncoding.EncodeToString(encryptedSecrets),
		Parameters:       store.Parameters,
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(encryptedStore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted store: %w", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write secrets file: %w", err)
	}

	return nil
}

// generateKeyFromPassword creates a 32-byte key from a password using SHA256
func generateKeyFromPassword(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

// encryptData encrypts data using AES-GCM
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

// decryptData decrypts data using AES-GCM
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

// getPassword retrieves the password from ANBUPW env var or prompts the user
func getPassword() (string, error) {
	// First check for environment variable
	password := os.Getenv("ANBUPW")
	if password != "" {
		return password, nil
	}

	// If not set, prompt the user
	fmt.Print("Enter password for secrets: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add a newline after input
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	return string(passwordBytes), nil
}

// ListSecrets lists all secret IDs and parameter IDs
func ListSecrets(filePath string) error {
	store, err := InitializeSecretsStore(filePath)
	if err != nil {
		return err
	}

	fmt.Println("\nSecrets:")
	if len(store.Secrets) == 0 {
		fmt.Println("  No secrets found")
	} else {
		for id := range store.Secrets {
			fmt.Printf("  %s\n", id)
		}
	}

	fmt.Println("\nParameters:")
	if len(store.Parameters) == 0 {
		fmt.Println("  No parameters found")
	} else {
		for id := range store.Parameters {
			fmt.Printf("  %s\n", id)
		}
	}

	return nil
}

// GetSecret retrieves a specific secret by ID
func GetSecret(filePath, secretID string) error {
	store, err := InitializeSecretsStore(filePath)
	if err != nil {
		return err
	}

	value, exists := store.Secrets[secretID]
	if !exists {
		return fmt.Errorf("secret '%s' not found", secretID)
	}

	fmt.Printf("%s: %s\n", secretID, value)
	return nil
}

// GetParameter retrieves a specific parameter by ID
func GetParameter(filePath, paramID string) error {
	store, err := InitializeSecretsStore(filePath)
	if err != nil {
		return err
	}

	value, exists := store.Parameters[paramID]
	if !exists {
		return fmt.Errorf("parameter '%s' not found", paramID)
	}

	fmt.Printf("%s: %s\n", paramID, value)
	return nil
}

// SetSecret sets a secret value (prompts for the value)
func SetSecret(filePath, secretID string) error {
	store, err := InitializeSecretsStore(filePath)
	if err != nil {
		return err
	}

	fmt.Printf("Enter value for secret '%s': ", secretID)
	value, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add a newline after input
	if err != nil {
		return fmt.Errorf("failed to read secret value: %w", err)
	}

	store.Secrets[secretID] = string(value)

	password, err := getPassword()
	if err != nil {
		return err
	}

	if err := saveSecretsStore(store, filePath, password); err != nil {
		return err
	}

	u.PrintSuccess(fmt.Sprintf("Secret '%s' set successfully", secretID))
	return nil
}

// SetParameter sets a parameter value (prompts for the value)
func SetParameter(filePath, paramID string) error {
	store, err := InitializeSecretsStore(filePath)
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

	password, err := getPassword()
	if err != nil {
		return err
	}

	if err := saveSecretsStore(store, filePath, password); err != nil {
		return err
	}

	u.PrintSuccess(fmt.Sprintf("Parameter '%s' set successfully", paramID))
	return nil
}

// DeleteSecret removes a secret by ID
func DeleteSecret(filePath, secretID string) error {
	store, err := InitializeSecretsStore(filePath)
	if err != nil {
		return err
	}

	if _, exists := store.Secrets[secretID]; !exists {
		return fmt.Errorf("secret '%s' not found", secretID)
	}

	delete(store.Secrets, secretID)

	password, err := getPassword()
	if err != nil {
		return err
	}

	if err := saveSecretsStore(store, filePath, password); err != nil {
		return err
	}

	u.PrintSuccess(fmt.Sprintf("Secret '%s' deleted successfully", secretID))
	return nil
}

// DeleteParameter removes a parameter by ID
func DeleteParameter(filePath, paramID string) error {
	store, err := InitializeSecretsStore(filePath)
	if err != nil {
		return err
	}

	if _, exists := store.Parameters[paramID]; !exists {
		return fmt.Errorf("parameter '%s' not found", paramID)
	}

	delete(store.Parameters, paramID)

	password, err := getPassword()
	if err != nil {
		return err
	}

	if err := saveSecretsStore(store, filePath, password); err != nil {
		return err
	}

	u.PrintSuccess(fmt.Sprintf("Parameter '%s' deleted successfully", paramID))
	return nil
}

// ImportSecrets imports secrets and parameters from a JSON file
func ImportSecrets(filePath, importFilePath string) error {
	importData, err := os.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	var importStore SecretsStore
	if err := json.Unmarshal(importData, &importStore); err != nil {
		return fmt.Errorf("failed to parse import file: %w", err)
	}

	store, err := InitializeSecretsStore(filePath)
	if err != nil {
		return err
	}

	// Merge imported data
	for id, value := range importStore.Secrets {
		store.Secrets[id] = value
	}

	for id, value := range importStore.Parameters {
		store.Parameters[id] = value
	}

	password, err := getPassword()
	if err != nil {
		return err
	}

	if err := saveSecretsStore(store, filePath, password); err != nil {
		return err
	}

	u.PrintSuccess(fmt.Sprintf("Imported %d secrets and %d parameters successfully",
		len(importStore.Secrets), len(importStore.Parameters)))
	return nil
}

// ExportSecrets exports secrets and parameters to a JSON file (unencrypted)
func ExportSecrets(filePath, exportFilePath string) error {
	store, err := InitializeSecretsStore(filePath)
	if err != nil {
		return err
	}

	exportData, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal export data: %w", err)
	}

	if err := os.WriteFile(exportFilePath, exportData, 0600); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	u.PrintSuccess(fmt.Sprintf("Exported %d secrets and %d parameters to %s",
		len(store.Secrets), len(store.Parameters), exportFilePath))
	return nil
}
