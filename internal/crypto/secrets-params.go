package anbuCrypto

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	u "github.com/tanq16/anbu/utils"
	"golang.org/x/term"
)

type SecretsStore struct {
	Secrets map[string]string `json:"secrets"`
}

type APIRequest struct {
	Command string `json:"command"`
	ID      string `json:"id,omitempty"`
	Value   string `json:"value,omitempty"`
	Content string `json:"content,omitempty"`
}

type APIResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message,omitempty"`
	Data    []string `json:"data,omitempty"`
	Value   string   `json:"value,omitempty"`
	Content string   `json:"content,omitempty"`
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

func GetPassword() (string, error) {
	password := os.Getenv("ANBUPW")
	if password != "" {
		return password, nil
	}
	fmt.Print("Enter password for secrets: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return string(passwordBytes), nil
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

func ImportSecrets(filePath, importFilePath string) error {
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
	password, err := GetPassword()
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

func ExportSecrets(filePath, exportFilePath string) error {
	store, err := loadSecretsStore(filePath)
	if err != nil {
		return err
	}
	exportStore := &SecretsStore{
		Secrets: make(map[string]string),
	}
	password, err := GetPassword()
	if err != nil {
		return err
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

func loggingHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u.PrintStream(fmt.Sprintf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path))
		h.ServeHTTP(w, r)
	}
}

func ServeSecrets(filePath string) error {
	http.HandleFunc("/list", loggingHandler(listHandler(filePath)))
	http.HandleFunc("/get", loggingHandler(getHandler(filePath)))
	http.HandleFunc("/add", loggingHandler(addHandler(filePath)))
	http.HandleFunc("/delete", loggingHandler(deleteHandler(filePath)))
	addr := ":8080"
	u.PrintSuccess(fmt.Sprintf("secrets server listening on %s", addr))
	return http.ListenAndServe(addr, nil)
}

func listHandler(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		secrets, err := ListSecrets(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := APIResponse{Status: "success", Data: secrets}
		json.NewEncoder(w).Encode(resp)
	}
}

func getHandler(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req APIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		password, err := GetPassword()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		value, err := GetSecret(filePath, req.ID, password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		resp := APIResponse{Status: "success", Value: value}
		json.NewEncoder(w).Encode(resp)
	}
}

func addHandler(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req APIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		password, err := GetPassword()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := SetSecret(filePath, req.ID, req.Value, password); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := APIResponse{Status: "success", Message: fmt.Sprintf("secret '%s' added", req.ID)}
		json.NewEncoder(w).Encode(resp)
	}
}

func deleteHandler(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req APIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := DeleteSecret(filePath, req.ID); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		resp := APIResponse{Status: "success", Message: fmt.Sprintf("secret '%s' deleted", req.ID)}
		json.NewEncoder(w).Encode(resp)
	}
}

// Client-side Remote Functions
func RemoteCall(host, command string, data map[string]string) ([]byte, error) {
	requestData := APIRequest{Command: command}
	if id, ok := data["id"]; ok {
		requestData.ID = id
	}
	if value, ok := data["value"]; ok {
		requestData.Value = value
	}
	if content, ok := data["content"]; ok {
		requestData.Content = content
	}
	jsonData, _ := json.Marshal(requestData)
	resp, err := http.Post(host, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func RemoteListSecrets(host string) error {
	body, err := RemoteCall(host+"/list", "list", nil)
	if err != nil {
		return err
	}
	var resp APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}
	if resp.Status != "success" {
		return fmt.Errorf("server error: %s", resp.Message)
	}
	u.PrintSuccess("Remote secrets:")
	for i, id := range resp.Data {
		fmt.Printf("  %d. %s\n", i+1, u.FInfo(id))
	}
	return nil
}

func RemoteGetSecret(host, secretID string) error {
	body, err := RemoteCall(host+"/get", "get", map[string]string{"id": secretID})
	if err != nil {
		return err
	}
	var resp APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}
	if resp.Status != "success" {
		return fmt.Errorf("server error: %s", resp.Message)
	}
	fmt.Println(resp.Value)
	return nil
}

func RemoteSetSecret(host, secretID, value string) error {
	body, err := RemoteCall(host+"/add", "add", map[string]string{"id": secretID, "value": value})
	if err != nil {
		return err
	}
	var resp APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}
	if resp.Status != "success" {
		return fmt.Errorf("server error: %s", resp.Message)
	}
	u.PrintSuccess(fmt.Sprintf("Secret '%s' set on remote server", secretID))
	return nil
}

func RemoteDeleteSecret(host, secretID string) error {
	body, err := RemoteCall(host+"/delete", "delete", map[string]string{"id": secretID})
	if err != nil {
		return err
	}
	var resp APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}
	if resp.Status != "success" {
		return fmt.Errorf("server error: %s", resp.Message)
	}
	u.PrintSuccess(fmt.Sprintf("Secret '%s' deleted from remote server", secretID))
	return nil
}

func ReadMultilineInput() (string, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter value (type 'EOF' on a new line to finish):")
	for scanner.Scan() {
		line := scanner.Text()
		if line == "EOF" {
			break
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	text := sb.String()
	if len(text) > 0 {
		text = text[:len(text)-1]
	}
	return text, nil
}

func ReadSingleLineInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	return strings.TrimSpace(text), nil
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
