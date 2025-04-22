package anbuNetwork

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tanq16/anbu/utils"
)

type HTTPServerOptions struct {
	ListenAddress string
	EnableUpload  bool
	EnableTLS     bool
	Domain        string
}

type HTTPServer struct {
	Options *HTTPServerOptions
	Server  *http.Server
}

func (s *HTTPServer) Start() error {
	fileServer := http.FileServer(http.Dir("."))

	// Setup handler with middlewares
	var handler http.Handler = fileServer
	if s.Options.EnableUpload {
		handler = s.uploadMiddleware(handler)
	}
	handler = s.loggingMiddleware(handler)

	// Setup server & TLS
	s.Server = &http.Server{
		Addr:    s.Options.ListenAddress,
		Handler: handler,
	}
	if s.Options.EnableTLS {
		tlsConfig, err := s.getTLSConfig()
		if err != nil {
			return err
		}
		s.Server.TLSConfig = tlsConfig
		fmt.Println(utils.OutSuccess(fmt.Sprintf("HTTPS server started on https://%s/", s.Options.ListenAddress)))
		return s.Server.ListenAndServeTLS("", "")
	}
	fmt.Println(utils.OutSuccess(fmt.Sprintf("HTTP server started on http://%s/", s.Options.ListenAddress)))
	return s.Server.ListenAndServe()
}

func (s *HTTPServer) Stop() error {
	if s.Server != nil {
		return s.Server.Close()
	}
	return nil
}

func (s *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(utils.OutDebug(fmt.Sprintf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)))
		next.ServeHTTP(w, r)
	})
}

func (s *HTTPServer) uploadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			targetPath := filepath.Join(".", r.URL.Path)
			// Ensure the path is within the server folder (prevent directory traversal)
			absTargetPath, _ := filepath.Abs(targetPath)
			serverRoot, _ := filepath.Abs(".")
			if !strings.HasPrefix(absTargetPath, serverRoot) {
				logger.Debug().Str("target", targetPath).Msg("attempted directory traversal")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			targetDir := filepath.Dir(targetPath)
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				logger.Debug().Err(err).Msg("failed to create directory")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			file, err := os.Create(targetPath)
			if err != nil {
				logger.Debug().Err(err).Msg("failed to create file")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			_, err = io.Copy(file, r.Body)
			if err != nil {
				logger.Debug().Err(err).Msg("failed to write file")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			fmt.Println(utils.OutSuccess(fmt.Sprintf("File uploaded to %s", targetPath)))
			w.WriteHeader(http.StatusCreated)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *HTTPServer) getTLSConfig() (*tls.Config, error) {
	cert, err := s.generateSelfSignedCert()
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed certificate: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

func (s *HTTPServer) generateSelfSignedCert() (tls.Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	domain := s.Options.Domain
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // 1 year validity
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

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	// Parse PEM to create tls.Certificate
	cert, err := tls.X509KeyPair(certPEM, privateKeyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}
	logger.Debug().Msg("generated self-signed certificate")
	return cert, nil
}
