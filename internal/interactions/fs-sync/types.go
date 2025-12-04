package fssync

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

type ManifestResponse struct {
	Files map[string]string `json:"files"`
}

type FileRequest struct {
	Paths []string `json:"paths"`
}

type FilesResponse struct {
	Files []FileContent `json:"files"`
}

type FileContent struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

type PathIgnorer struct {
	patterns []string
}

func NewPathIgnorer(ignoreStr string) *PathIgnorer {
	if ignoreStr == "" {
		return &PathIgnorer{patterns: []string{}}
	}
	parts := strings.Split(ignoreStr, ",")
	patterns := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			patterns = append(patterns, trimmed)
		}
	}
	return &PathIgnorer{patterns: patterns}
}

func (pi *PathIgnorer) IsIgnored(path string) bool {
	for _, pattern := range pi.patterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func ComputeFileHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func BuildManifest(rootDir string, ignorer *PathIgnorer) (map[string]string, error) {
	manifest := make(map[string]string)
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		relPath = filepath.ToSlash(relPath)
		if ignorer.IsIgnored(relPath) {
			return nil
		}
		hash, err := ComputeFileHash(path)
		if err != nil {
			return err
		}
		manifest[relPath] = hash
		return nil
	})
	return manifest, err
}
