package anbuNetwork

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func NewHTTPServer(options *HTTPServerOptions) *HTTPServer {
	return &HTTPServer{
		Options: options,
	}
}

func (s *HTTPServer) Setup() error {
	var handler http.Handler
	if s.Options.EnableUpload {
		handler = http.HandlerFunc(s.handleUpload)
	} else {
		handler = http.FileServer(http.Dir("."))
	}
	s.Server = &http.Server{
		Addr:    s.Options.ListenAddress,
		Handler: withHTTPLogging(handler),
	}
	if s.Options.EnableTLS {
		tlsConfig, err := s.getTLSConfig()
		if err != nil {
			return err
		}
		s.Server.TLSConfig = tlsConfig
	}
	return nil
}

func (s *HTTPServer) Run() error {
	if s.Options.EnableTLS {
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

func withHTTPLogging(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u.PrintStream(fmt.Sprintf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path))
		next.ServeHTTP(w, r)
	}
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
		reader, err := r.MultipartReader()
		if err != nil {
			u.PrintError("failed to get multipart reader", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				u.PrintError("failed to read multipart part", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			switch part.FormName() {
			case "text":
				filename := fmt.Sprintf("text-%d.txt", time.Now().Unix())
				path, n, err := streamPartToUniqueFile(filename, part)
				if err != nil {
					u.PrintError("failed to write text file", err)
					continue
				}
				if n == 0 {
					os.Remove(path)
					continue
				}
				u.PrintInfo(fmt.Sprintf("Text saved to %s", path))
			case "files":
				filename := part.FileName()
				if filename == "" {
					continue
				}
				path, _, err := streamPartToUniqueFile(filename, part)
				if err != nil {
					u.PrintError("failed to write file", err)
					continue
				}
				u.PrintInfo(fmt.Sprintf("File uploaded to %s", path))
			}
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

func sanitizeUploadFilename(filename string) string {
	filename = filepath.Base(filename)
	if filename == "." || filename == ".." || filename == "/" || filename == string(filepath.Separator) {
		return "uploaded_file"
	}
	return filename
}

func createUniqueFile(filename string) (*os.File, string, error) {
	filename = sanitizeUploadFilename(filename)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	for counter := 0; ; counter++ {
		candidate := filename
		if counter > 0 {
			candidate = fmt.Sprintf("%s-%d%s", name, counter, ext)
		}
		f, err := os.OpenFile(candidate, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err == nil {
			return f, candidate, nil
		}
		if !os.IsExist(err) {
			return nil, "", err
		}
	}
}

func streamPartToUniqueFile(filename string, part io.Reader) (string, int64, error) {
	outFile, path, err := createUniqueFile(filename)
	if err != nil {
		return "", 0, err
	}
	n, copyErr := io.Copy(outFile, part)
	closeErr := outFile.Close()
	if copyErr != nil {
		os.Remove(path)
		return path, n, copyErr
	}
	if closeErr != nil {
		os.Remove(path)
		return path, n, closeErr
	}
	return path, n, nil
}

func (s *HTTPServer) getTLSConfig() (*tls.Config, error) {
	cert, err := u.GenerateSelfSignedCert()
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed certificate: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}
