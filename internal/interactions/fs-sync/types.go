package fssync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type MessageType string

const (
	TypeManifest      MessageType = "manifest"
	TypeFileRequest   MessageType = "file_request"
	TypeFileContent   MessageType = "file_content"
	TypeFileOperation MessageType = "file_operation"
)

type OperationType string

const (
	OpWrite  OperationType = "write"
	OpRemove OperationType = "remove"
)

type MessageWrapper struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ManifestMessage struct {
	Files map[string]string `json:"files"`
}

type FileRequestMessage struct {
	Paths []string `json:"paths"`
}

type FileContentMessage struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

type FileOperationMessage struct {
	Op      OperationType `json:"op"`
	Path    string        `json:"path"`
	Content []byte        `json:"content"`
	IsDir   bool          `json:"is_dir,omitempty"`
}

type SafeConn struct {
	Conn *websocket.Conn
	mu   sync.Mutex
}

func (sc *SafeConn) WriteJSON(v any) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.Conn.WriteJSON(v)
}

type PathIgnorer struct {
	patterns []string
}

func NewPathIgnorer(ignoreStr string) *PathIgnorer {
	if ignoreStr == "" {
		return &PathIgnorer{patterns: []string{}}
	}
	return &PathIgnorer{patterns: strings.Split(ignoreStr, ",")}
}

func (pi *PathIgnorer) IsIgnored(path string) bool {
	for _, pattern := range pi.patterns {
		match, err := filepath.Match(pattern, path)
		if err == nil && match {
			return true
		}
	}
	return false
}

func ComputeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func BuildFileManifest(rootDir string, ignorer *PathIgnorer) (map[string]string, error) {
	manifest := make(map[string]string)
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		if ignorer.IsIgnored(relPath) {
			log.Debug().Msgf("Ignoring path based on rules: %s", relPath)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			hash, err := ComputeFileHash(path)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to compute hash for file: %s", path)
				return nil
			}
			manifest[relPath] = hash
		}
		return nil
	})
	return manifest, err
}

func ApplyOperation(rootDir string, op *FileOperationMessage) error {
	fullPath := filepath.Join(rootDir, op.Path)
	switch op.Op {
	case OpWrite:
		if op.IsDir {
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				log.Error().Err(err).Msgf("Failed to create directory: %s", fullPath)
				return err
			}
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			log.Error().Err(err).Msgf("Failed to create parent directories: %s", fullPath)
			return err
		}
		if err := os.WriteFile(fullPath, op.Content, 0644); err != nil {
			log.Error().Err(err).Msgf("Failed to write file: %s", fullPath)
			return err
		}
	case OpRemove:
		if err := os.RemoveAll(fullPath); err != nil {
			log.Error().Err(err).Msgf("Failed to remove file/directory: %s", fullPath)
			return err
		}
	}
	return nil
}
