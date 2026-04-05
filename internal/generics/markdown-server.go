package anbuGenerics

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	u "github.com/tanq16/anbu/utils"
)

//go:embed markdown-viewer.html
var markdownViewerHTML []byte

//go:embed static
var staticFiles embed.FS

type FileNode struct {
	IsDir    bool                `json:"isDir,omitempty"`
	Children map[string]FileNode `json:"children,omitempty"`
}

type MarkdownServerOptions struct {
	ListenAddress string
	RootDir       string
}

type MarkdownServer struct {
	Options *MarkdownServerOptions
	mux     *http.ServeMux
}

func NewMarkdownServer(listenAddr string) (*MarkdownServer, error) {
	rootDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &MarkdownServer{
		Options: &MarkdownServerOptions{
			ListenAddress: listenAddr,
			RootDir:       rootDir,
		},
	}, nil
}

func (s *MarkdownServer) Setup() error {
	s.mux = http.NewServeMux()
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return err
	}
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	s.mux.HandleFunc("/", s.serveHTML)
	s.mux.HandleFunc("/api/tree", s.serveFileTree)
	s.mux.HandleFunc("/api/blob", s.serveFileContent)
	s.mux.HandleFunc("/files/", s.serveRawFile)
	return nil
}

func (s *MarkdownServer) Run() error {
	u.PrintInfo(fmt.Sprintf("Markdown viewer started at http://%s/", s.Options.ListenAddress))
	return http.ListenAndServe(s.Options.ListenAddress, withLogging(s.mux))
}

func StartMarkdownServer(listenAddr string) error {
	server, err := NewMarkdownServer(listenAddr)
	if err != nil {
		return err
	}
	if err := server.Setup(); err != nil {
		return err
	}
	return server.Run()
}

func (s *MarkdownServer) serveHTML(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(markdownViewerHTML)
}

func (s *MarkdownServer) serveFileTree(w http.ResponseWriter, r *http.Request) {
	tree, err := s.buildFileTree(s.Options.RootDir)
	if err != nil {
		u.PrintError("failed to build file tree", err)
		http.Error(w, "Failed to build file tree", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree)
}

var supportedExts = map[string]bool{
	".md": true, ".markdown": true, ".txt": true,
}

var imageContentTypes = map[string]string{
	".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg",
	".gif": "image/gif", ".svg": "image/svg+xml", ".webp": "image/webp",
	".ico": "image/x-icon", ".bmp": "image/bmp",
}

func (s *MarkdownServer) serveFileContent(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Missing path parameter", http.StatusBadRequest)
		return
	}
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(s.Options.RootDir, cleanPath)
	if !strings.HasPrefix(fullPath, s.Options.RootDir) {
		http.Error(w, "Invalid path", http.StatusForbidden)
		return
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error accessing file", http.StatusInternalServerError)
		}
		return
	}
	if info.IsDir() {
		http.Error(w, "Path is a directory", http.StatusBadRequest)
		return
	}
	ext := strings.ToLower(filepath.Ext(fullPath))
	if !supportedExts[ext] {
		http.Error(w, "Unable to Render", http.StatusUnsupportedMediaType)
		return
	}
	content, err := os.ReadFile(fullPath)
	if err != nil {
		u.PrintError("failed to read file", err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(content)
}

func (s *MarkdownServer) serveRawFile(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/files/")
	if path == "" {
		http.Error(w, "Missing path", http.StatusBadRequest)
		return
	}
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(s.Options.RootDir, cleanPath)
	if !strings.HasPrefix(fullPath, s.Options.RootDir) {
		http.Error(w, "Invalid path", http.StatusForbidden)
		return
	}
	ext := strings.ToLower(filepath.Ext(fullPath))
	contentType, ok := imageContentTypes[ext]
	if !ok {
		http.Error(w, "Not an image file", http.StatusUnsupportedMediaType)
		return
	}
	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(content)
}

func (s *MarkdownServer) buildFileTree(rootPath string) (map[string]FileNode, error) {
	tree := make(map[string]FileNode)
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(d.Name(), ".") && path != rootPath {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(d.Name()))
			if !supportedExts[ext] {
				return nil
			}
		}
		parts := strings.Split(relPath, string(filepath.Separator))
		current := tree
		for i, part := range parts {
			isLast := i == len(parts)-1
			if isLast {
				if d.IsDir() {
					if _, exists := current[part]; !exists {
						current[part] = FileNode{
							IsDir:    true,
							Children: make(map[string]FileNode),
						}
					}
				} else {
					current[part] = FileNode{
						IsDir: false,
					}
				}
			} else {
				if _, exists := current[part]; !exists {
					current[part] = FileNode{
						IsDir:    true,
						Children: make(map[string]FileNode),
					}
				}
				node := current[part]
				current = node.Children
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	pruneEmptyDirs(tree)
	return tree, nil
}

func pruneEmptyDirs(tree map[string]FileNode) {
	for name, node := range tree {
		if node.IsDir {
			pruneEmptyDirs(node.Children)
			if len(node.Children) == 0 {
				delete(tree, name)
			}
		}
	}
}

func withLogging(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u.PrintStream(fmt.Sprintf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path))
		next.ServeHTTP(w, r)
	}
}
