package gdrive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	u "github.com/tanq16/anbu/utils"
	"google.golang.org/api/drive/v3"
)

type IndexItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	ModifiedTime string `json:"modified_time"`
}

type IndexStore struct {
	Provider  string      `json:"provider"`
	RootPath  string      `json:"root_path"`
	Timestamp time.Time   `json:"timestamp"`
	Items     []IndexItem `json:"items"`
}

type node struct {
	ID       string
	Name     string
	ParentID string
	MimeType string
	Size     int64
	ModTime  string
	FullPath string
}

func getIndexPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".anbu-gdrive-index.json"), nil
}

func saveIndex(rootPath string, items []IndexItem) error {
	path, err := getIndexPath()
	if err != nil {
		return err
	}
	store := IndexStore{
		Provider:  "gdrive",
		RootPath:  rootPath,
		Timestamp: time.Now(),
		Items:     items,
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func loadIndex() (*IndexStore, error) {
	path, err := getIndexPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var store IndexStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return &store, nil
}

func SearchIndex(pattern string, searchPath string, excludeDirs, excludeFiles bool) ([]IndexItem, error) {
	idx, err := loadIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}
	cleanSearch := strings.Trim(searchPath, "/")
	cleanRoot := strings.Trim(idx.RootPath, "/")
	isRootIndex := cleanRoot == "root" || cleanRoot == ""
	if cleanSearch != "" && !isRootIndex && !strings.HasPrefix(cleanSearch, cleanRoot) {
		return nil, fmt.Errorf("search path '%s' is outside the indexed root '%s'. Please run index command on this path first", searchPath, idx.RootPath)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}
	var results []IndexItem
	for _, item := range idx.Items {
		if excludeDirs && item.Type == "folder" {
			continue
		}
		if excludeFiles && item.Type != "folder" {
			continue
		}
		itemP := strings.Trim(item.Path, "/")
		if !strings.HasPrefix(itemP, cleanSearch) {
			continue
		}
		if re.MatchString(item.Name) {
			results = append(results, item)
		}
	}
	return results, nil
}

func GenerateIndex(srv *drive.Service, rootPath string) error {
	u.PrintInfo("Fetching file metadata (flat fetch)...")
	fileMap := make(map[string]*node)
	query := "trashed = false"
	pageToken := ""
	for {
		call := srv.Files.List().Q(query).Fields("nextPageToken, files(id, name, parents, mimeType, size, modifiedTime)").PageSize(1000)
		if pageToken != "" {
			call.PageToken(pageToken)
		}
		r, err := call.Do()
		if err != nil {
			return err
		}
		for _, f := range r.Files {
			pid := ""
			if len(f.Parents) > 0 {
				pid = f.Parents[0]
			}
			fileMap[f.Id] = &node{
				ID:       f.Id,
				Name:     f.Name,
				ParentID: pid,
				MimeType: f.MimeType,
				Size:     f.Size,
				ModTime:  f.ModifiedTime,
			}
		}
		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}
	u.PrintInfo(fmt.Sprintf("Fetched %d items. Reconstructing tree...", len(fileMap)))
	var resolve func(*node) string
	resolve = func(n *node) string {
		if n.FullPath != "" {
			return n.FullPath
		}
		parent, exists := fileMap[n.ParentID]
		if !exists || n.ParentID == "" {
			n.FullPath = n.Name
			return n.Name
		}
		n.FullPath = filepath.Join(resolve(parent), n.Name)
		return n.FullPath
	}
	var items []IndexItem
	targetRoot := rootPath
	if targetRoot == "root" {
		targetRoot = ""
	}
	targetRoot = strings.Trim(targetRoot, "/")
	for _, n := range fileMap {
		path := resolve(n)
		cleanPath := strings.Trim(path, "/")
		if targetRoot != "" && !strings.HasPrefix(cleanPath, targetRoot) {
			continue
		}
		itemType := "file"
		if n.MimeType == googleFolderMimeType {
			itemType = "folder"
		}
		items = append(items, IndexItem{
			ID:           n.ID,
			Name:         n.Name,
			Path:         path,
			Type:         itemType,
			Size:         n.Size,
			ModifiedTime: n.ModTime,
		})
	}
	return saveIndex(rootPath, items)
}
