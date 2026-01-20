package fssync

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	u "github.com/tanq16/anbu/utils"
)

type ServerConfig struct {
	Port        int
	SyncDir     string
	IgnorePaths string
	EnableTLS   bool
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
			u.PrintWarning(fmt.Sprintf("Invalid path: %s", path), nil)
			continue
		}
		fullPath := filepath.Join(s.cfg.SyncDir, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			u.PrintWarning(fmt.Sprintf("Failed to read file: %s", path), err)
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

func (s *Server) getTLSConfig() (*tls.Config, error) {
	cert, err := s.generateSelfSignedCert()
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed certificate: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

func (s *Server) generateSelfSignedCert() (tls.Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	domain := "localhost"
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Anbu Self-Signed Certificate"},
			CommonName:   domain,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain, "localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	cert, err := tls.X509KeyPair(certPEM, privateKeyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}
	return cert, nil
}
