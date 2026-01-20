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
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
)

type HTTPServerOptions struct {
	ListenAddress string
	EnableUpload  bool
	EnableTLS     bool
}

type HTTPServer struct {
	Options *HTTPServerOptions
	Server  *http.Server
}

func (s *HTTPServer) Start() error {
	var handler http.Handler
	if s.Options.EnableUpload {
		handler = http.HandlerFunc(s.handleUpload)
	} else {
		handler = http.FileServer(http.Dir("."))
	}
	handler = s.loggingMiddleware(handler)
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
		u.PrintInfo(fmt.Sprintf("HTTPS server started on https://%s/", s.Options.ListenAddress))
		return s.Server.ListenAndServeTLS("", "")
	}
	u.PrintInfo(fmt.Sprintf("HTTP server started on http://%s/", s.Options.ListenAddress))
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
		u.PrintStream(fmt.Sprintf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path))
		next.ServeHTTP(w, r)
	})
}

func (s *HTTPServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
<style>
body { background-color: #2a2a2a; color: #fff; font-family: sans-serif; padding: 20px; }
.container { max-width: 600px; margin: 0 auto; }
textarea { width: 100%; height: 300px; padding: 10px; background-color: #1a1a1a; color: #fff; border: 2px dashed #555; border-radius: 5px; box-sizing: border-box; }
textarea.drag-over { border-color: #4a9eff; background-color: #252525; }
input[type="file"] { margin: 15px 0; color: #fff; }
button { background-color: #4a9eff; color: #fff; padding: 10px 30px; border: none; border-radius: 5px; cursor: pointer; font-size: 16px; }
button:hover { background-color: #3a8eef; }
button:disabled { background-color: #555; cursor: not-allowed; }
.progress-container { display: none; margin: 15px 0; }
.progress-bar { width: 100%; height: 20px; background-color: #1a1a1a; border-radius: 10px; overflow: hidden; }
.progress-fill { height: 100%; background-color: #4a9eff; width: 0%; transition: width 0.1s; }
.progress-text { margin-top: 5px; font-size: 14px; color: #aaa; }
</style>
</head>
<body>
<div class="container">
<h2>Upload</h2>
<form method="POST" enctype="multipart/form-data" id="uploadForm">
<textarea name="text" id="textarea" placeholder="Paste or drag files here..."></textarea>
<input type="file" name="files" multiple id="fileInput">
<button type="submit" id="submitBtn">Upload</button>
<div class="progress-container" id="progressContainer">
<div class="progress-bar">
<div class="progress-fill" id="progressFill"></div>
</div>
<div class="progress-text" id="progressText">0%</div>
</div>
</form>
</div>
<script>
var textarea = document.getElementById('textarea');
var form = document.getElementById('uploadForm');
var submitBtn = document.getElementById('submitBtn');
var progressContainer = document.getElementById('progressContainer');
var progressFill = document.getElementById('progressFill');
var progressText = document.getElementById('progressText');

textarea.addEventListener('dragover', function(e) { e.preventDefault(); textarea.classList.add('drag-over'); });
textarea.addEventListener('dragleave', function(e) { e.preventDefault(); textarea.classList.remove('drag-over'); });
textarea.addEventListener('drop', function(e) {
  e.preventDefault();
  textarea.classList.remove('drag-over');
  var files = e.dataTransfer.files;
  if (files.length > 0) {
    var fileInput = document.querySelector('input[type="file"]');
    fileInput.files = files;
  }
});

form.addEventListener('submit', function(e) {
  e.preventDefault();
  var formData = new FormData(form);
  var xhr = new XMLHttpRequest();
  progressContainer.style.display = 'block';
  submitBtn.disabled = true;
  progressFill.style.width = '0%';
  progressText.textContent = '0%';
  xhr.upload.addEventListener('progress', function(e) {
    if (e.lengthComputable) {
      var percent = (e.loaded / e.total) * 100;
      progressFill.style.width = percent + '%';
      progressText.textContent = Math.round(percent) + '%';
    }
  });
  xhr.addEventListener('load', function() {
    if (xhr.status === 200) {
      document.body.innerHTML = xhr.responseText;
    } else {
      progressText.textContent = 'Upload failed';
      submitBtn.disabled = false;
    }
  });
  xhr.addEventListener('error', function() {
    progressText.textContent = 'Upload failed';
    submitBtn.disabled = false;
  });
  xhr.open('POST', '/');
  xhr.send(formData);
});
</script>
</body>
</html>`)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			u.PrintError("failed to parse multipart form")
			log.Debug().Err(err).Msg("failed to parse multipart form")
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		textContent := r.FormValue("text")
		if strings.TrimSpace(textContent) != "" {
			epoch := time.Now().Unix()
			filename := fmt.Sprintf("text-%d.txt", epoch)
			filename = s.ensureUniqueFilename(filename)
			if err := os.WriteFile(filename, []byte(textContent), 0644); err != nil {
				u.PrintError("failed to write text file")
				log.Debug().Err(err).Msg("failed to write text file")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			u.PrintInfo(fmt.Sprintf("Text saved to %s", filename))
		}

		files := r.MultipartForm.File["files"]
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				u.PrintError("failed to open uploaded file")
				log.Debug().Err(err).Msg("failed to open uploaded file")
				continue
			}
			filename := s.ensureUniqueFilename(fileHeader.Filename)
			outFile, err := os.Create(filename)
			if err != nil {
				u.PrintError("failed to create file")
				log.Debug().Err(err).Msg("failed to create file")
				file.Close()
				continue
			}
			_, err = io.Copy(outFile, file)
			file.Close()
			outFile.Close()
			if err != nil {
				u.PrintError("failed to write file")
				log.Debug().Err(err).Msg("failed to write file")
				continue
			}
			u.PrintInfo(fmt.Sprintf("File uploaded to %s", filename))
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
<style>
body { background-color: #2a2a2a; color: #fff; font-family: sans-serif; padding: 20px; text-align: center; }
</style>
</head>
<body>
<h2>Upload Successful</h2>
<p><a href="/" style="color: #4a9eff;">Upload more</a></p>
</body>
</html>`)
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (s *HTTPServer) ensureUniqueFilename(filename string) string {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return filename
	}
	epoch := time.Now().Unix()
	return fmt.Sprintf("%d-%s", epoch, filename)
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
	log.Debug().Msg("generated private key")
	domain := "localhost"
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
	log.Debug().Msg("created certificate")
	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	// Parse PEM to create tls.Certificate
	cert, err := tls.X509KeyPair(certPEM, privateKeyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}
	log.Debug().Msg("generated self-signed certificate")
	return cert, nil
}
