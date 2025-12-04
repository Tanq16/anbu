package fssync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type ServerConfig struct {
	Port        int
	SyncDir     string
	IgnorePaths string
	TLSCert     string
	TLSKey      string
}

type Server struct {
	cfg       ServerConfig
	ignorer   *PathIgnorer
	serveDone chan struct{}
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

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/manifest", s.handleManifest)
	mux.HandleFunc("/files", s.handleFiles)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Port),
		Handler: mux,
	}
	go func() {
		var err error
		if s.cfg.TLSCert != "" && s.cfg.TLSKey != "" {
			err = server.ListenAndServeTLS(s.cfg.TLSCert, s.cfg.TLSKey)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Server error")
		}
	}()
	<-s.serveDone
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	return nil
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
			log.Warn().Msgf("Invalid path: %s", path)
			continue
		}
		fullPath := filepath.Join(s.cfg.SyncDir, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to read file: %s", path)
			continue
		}
		files = append(files, FileContent{
			Path:    path,
			Content: content,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FilesResponse{Files: files})
	select {
	case <-s.serveDone:
	default:
		close(s.serveDone)
	}
}
