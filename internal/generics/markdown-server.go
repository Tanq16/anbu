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

	u "github.com/tanq16/anbu/internal/utils"
)

//go:embed markdown-viewer.html
var markdownViewerHTML []byte

//go:embed static/*
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
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return err
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.HandleFunc("/", server.serveHTML)
	mux.HandleFunc("/api/tree", server.serveFileTree)
	mux.HandleFunc("/api/blob", server.serveFileContent)
	u.PrintInfo(fmt.Sprintf("Markdown viewer started at http://%s/", listenAddr))
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
		u.PrintError("failed to build file tree", err)
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
		u.PrintError("failed to read file", err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	ext := strings.ToLower(filepath.Ext(fullPath))
	filename := filepath.Base(fullPath)
	if ext == ".md" || ext == ".markdown" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(content)
		return
	}
	lang := getLanguageFromFilename(filename)
	if lang == "" {
		if ext == "" {
			http.Error(w, "File type not supported for rendering", http.StatusUnsupportedMediaType)
			return
		}
		lang = getLanguageFromExtension(ext)
		if lang == "" {
			http.Error(w, "File type not supported for rendering", http.StatusUnsupportedMediaType)
			return
		}
	}
	content = fmt.Appendf(nil, "```%s\n%s\n```", lang, string(content))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(content)
}

func getLanguageFromExtension(ext string) string {
	ext = strings.TrimPrefix(ext, ".")

	// Thanks AI
	langMap := map[string]string{
		"go": "go",
		"py": "python", "pyw": "python", "pyi": "python",
		"js": "javascript", "jsx": "javascript", "mjs": "javascript", "cjs": "javascript",
		"ts": "typescript", "tsx": "typescript",
		"java": "java",
		"c":    "c", "h": "c",
		"cpp": "cpp", "cc": "cpp", "cxx": "cpp", "hpp": "cpp", "hxx": "cpp",
		"rs": "rust",
		"rb": "ruby", "rake": "ruby",
		"php": "php", "phtml": "php",
		"sh": "bash", "bash": "bash", "zsh": "bash",
		"yaml": "yaml", "yml": "yaml",
		"json": "json",
		"xml":  "xml", "html": "html", "htm": "html",
		"css": "css", "scss": "scss", "sass": "sass", "less": "less",
		"sql":   "sql",
		"swift": "swift",
		"kt":    "kotlin", "kts": "kotlin",
		"hs": "haskell", "lhs": "haskell",
		"ml": "ocaml", "mli": "ocaml",
		"lua": "lua",
		"r":   "r",
		"pl":  "perl", "pm": "perl",
		"vim":        "vim",
		"dockerfile": "dockerfile",
		"makefile":   "makefile", "mk": "makefile",
		"cmake":      "cmake",
		"toml":       "toml",
		"ini":        "ini",
		"properties": "properties",
		"diff":       "diff", "patch": "diff",
		"tex":  "latex",
		"dart": "dart",
		"cs":   "csharp",
		"ps1":  "powershell", "psm1": "powershell", "psd1": "powershell",
		"tf":      "hcl",
		"hcl":     "hcl",
		"graphql": "graphql", "gql": "graphql",
		"proto": "protobuf",
		"awk":   "awk",
		"sed":   "sed",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return ""
}

func getLanguageFromFilename(filename string) string {
	filenameMap := map[string]string{
		"LICENSE":         "text",
		"Dockerfile":      "dockerfile",
		"Makefile":        "makefile",
		"CMakeLists.txt":  "cmake",
		"Rakefile":        "ruby",
		"Gemfile":         "ruby",
		"Gemfile.lock":    "ruby",
		"BUILD":           "python",
		"BUILD.bazel":     "python",
		"WORKSPACE":       "python",
		"WORKSPACE.bazel": "python",
		"Justfile":        "makefile",
		"justfile":        "makefile",
	}
	if lang, ok := filenameMap[filename]; ok {
		return lang
	}
	if lang, ok := filenameMap[strings.ToLower(filename)]; ok {
		return lang
	}
	return ""
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
		u.PrintStream(fmt.Sprintf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path))
		next.ServeHTTP(w, r)
	})
}
