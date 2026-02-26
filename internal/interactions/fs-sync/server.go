package fssync

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	u "github.com/tanq16/anbu/internal/utils"
)

type ServerConfig struct {
	Port        int
	SyncDir     string
	IgnorePaths string
	EnableTLS   bool
	Mode        string
	DeleteExtra bool
	DryRun      bool
}

type Server struct {
	cfg            ServerConfig
	ignorer        *PathIgnorer
	serveDone      chan struct{}
	closeOnce      sync.Once
	senderManifest map[string]string
}

func NewServer(cfg ServerConfig) (*Server, error) {
	absDir, err := filepath.Abs(cfg.SyncDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve sync directory: %w", err)
	}
	cfg.SyncDir = absDir
	if _, err := os.Stat(cfg.SyncDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("sync directory does not exist: %s", cfg.SyncDir)
	}
	return &Server{
		cfg:       cfg,
		ignorer:   NewPathIgnorer(cfg.IgnorePaths),
		serveDone: make(chan struct{}),
	}, nil
}

func (s *Server) shutdown() {
	s.closeOnce.Do(func() { close(s.serveDone) })
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mode", s.handleMode)
	mux.HandleFunc("/manifest", s.handleManifest)
	if s.cfg.Mode == "send" {
		mux.HandleFunc("/files", s.handleFiles)
	} else {
		mux.HandleFunc("/upload", s.handleUpload)
	}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Port),
		Handler: mux,
	}
	if s.cfg.EnableTLS {
		tlsConfig, err := s.getTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to get TLS config: %w", err)
		}
		server.TLSConfig = tlsConfig
	}
	go func() {
		var err error
		if s.cfg.EnableTLS {
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			u.PrintError("Server error", err)
		}
	}()
	<-s.serveDone
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	return nil
}

func (s *Server) handleMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ModeResponse{Mode: s.cfg.Mode})
}

func (s *Server) handleManifest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	manifest, err := BuildManifest(s.cfg.SyncDir, s.ignorer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ManifestResponse{Files: manifest})
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var files []FileContent
	for _, path := range req.Paths {
		if s.ignorer.IsIgnored(path) {
			continue
		}
		relPath := filepath.Clean(path)
		if strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
			u.PrintWarn(fmt.Sprintf("Invalid path: %s", path), nil)
			continue
		}
		fullPath := filepath.Join(s.cfg.SyncDir, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			u.PrintWarn(fmt.Sprintf("Failed to read file: %s", path), err)
			continue
		}
		files = append(files, FileContent{
			Path:    path,
			Content: content,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FilesResponse{Files: files})
	s.shutdown()
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var uploadReq UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&uploadReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if s.cfg.DryRun {
		for _, file := range uploadReq.Files {
			u.PrintGeneric(fmt.Sprintf("Dry Run: %s", u.FDebug(file.Path)))
		}
		if s.cfg.DeleteExtra {
			for _, path := range uploadReq.ToDelete {
				u.PrintGeneric(fmt.Sprintf("Dry Run (delete): %s", u.FDebug(path)))
			}
		}
		u.LineBreak()
		totalCount := len(uploadReq.Files) + len(uploadReq.ToDelete)
		if totalCount == 0 {
			u.PrintWarn("no files would be synced", nil)
		} else {
			u.PrintGeneric(fmt.Sprintf("%s %s", u.FDebug("Operation completed:"), u.FSuccess(fmt.Sprintf("%d file(s) would be synced", totalCount))))
		}
		w.WriteHeader(http.StatusOK)
		s.shutdown()
		return
	}

	count := 0
	for _, file := range uploadReq.Files {
		relPath := filepath.Clean(file.Path)
		if strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
			u.PrintWarn(fmt.Sprintf("Invalid path: %s", file.Path), nil)
			continue
		}
		fullPath := filepath.Join(s.cfg.SyncDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			u.PrintError(fmt.Sprintf("Failed to create directory for %s", file.Path), err)
			continue
		}
		if err := os.WriteFile(fullPath, file.Content, 0644); err != nil {
			u.PrintError(fmt.Sprintf("Failed to write %s", file.Path), err)
			continue
		}
		u.PrintGeneric(fmt.Sprintf("Received: %s", u.FSuccess(file.Path)))
		count++
	}

	deletedCount := 0
	if s.cfg.DeleteExtra {
		for _, path := range uploadReq.ToDelete {
			relPath := filepath.Clean(path)
			if strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
				continue
			}
			fullPath := filepath.Join(s.cfg.SyncDir, relPath)
			if err := os.RemoveAll(fullPath); err != nil {
				u.PrintError(fmt.Sprintf("Failed to delete %s", path), err)
			} else {
				u.PrintGeneric(fmt.Sprintf("Deleted: %s", u.FSuccess(path)))
				deletedCount++
			}
		}
	}

	u.LineBreak()
	totalCount := count + deletedCount
	if totalCount == 0 {
		u.PrintWarn("no files were synced", nil)
	} else {
		u.PrintGeneric(fmt.Sprintf("%s %s", u.FDebug("Operation completed:"), u.FSuccess(fmt.Sprintf("%d file(s) synced", totalCount))))
	}
	w.WriteHeader(http.StatusOK)
	s.shutdown()
}

func (s *Server) getTLSConfig() (*tls.Config, error) {
	cert, err := u.GenerateSelfSignedCert()
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed certificate: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}
