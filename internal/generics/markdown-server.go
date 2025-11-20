package anbuGenerics

import (
	_ "embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:embed markdown-viewer.html
var markdownViewerHTML []byte

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
}

func StartMarkdownServer(listenAddr string) error {
	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}
	server := &MarkdownServer{
		Options: &MarkdownServerOptions{
			ListenAddress: listenAddr,
			RootDir:       rootDir,
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.serveHTML)
	mux.HandleFunc("/api/tree", server.serveFileTree)
	mux.HandleFunc("/api/blob", server.serveFileContent)
	log.Info().Msgf("Markdown viewer started at http://%s/", listenAddr)
	return http.ListenAndServe(listenAddr, loggingMiddleware(mux))
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
		log.Error().Err(err).Msg("Failed to build file tree")
		http.Error(w, "Failed to build file tree", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree)
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
	content, err := os.ReadFile(fullPath)
	if err != nil {
		log.Error().Err(err).Str("path", fullPath).Msg("Failed to read file")
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
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
	return tree, err
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msgf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
