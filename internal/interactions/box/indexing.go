package box

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	u "github.com/tanq16/anbu/utils"
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

func getIndexPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	anbuDir := filepath.Join(home, ".anbu")
	if err := os.MkdirAll(anbuDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(anbuDir, "box-index.json"), nil
}

func saveIndex(rootPath string, items []IndexItem) error {
	path, err := getIndexPath()
	if err != nil {
		return err
	}
	store := IndexStore{
		Provider:  "box",
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
	if cleanSearch == "" || cleanSearch == "/" {
		cleanSearch = ""
	}
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
		if cleanSearch != "" && !strings.HasPrefix(itemP, cleanSearch) {
			continue
		}
		if re.MatchString(item.Name) {
			results = append(results, item)
		}
	}
	return results, nil
}

func GenerateIndex(client *http.Client, rootPath string) error {
	folderID, _, err := resolvePathToID(client, rootPath, "folder")
	if err != nil {
		return err
	}
	var items []IndexItem
	u.PrintInfo("Crawling Box directory structure...")
	var crawl func(fid, currentPath string) error
	crawl = func(fid, currentPath string) error {
		folders, files, err := listBoxFolderContents(client, fid)
		if err != nil {
			return err
		}
		for _, f := range files {
			fullP := filepath.Join(currentPath, f.Name)
			items = append(items, IndexItem{
				ID:           f.ID,
				Name:         f.Name,
				Path:         fullP,
				Type:         "file",
				Size:         f.Size,
				ModifiedTime: f.ModifiedTime,
			})
		}
		for _, f := range folders {
			fullP := filepath.Join(currentPath, f.Name)
			items = append(items, IndexItem{
				ID:           f.ID,
				Name:         f.Name,
				Path:         fullP,
				Type:         "folder",
				ModifiedTime: f.ModifiedTime,
			})
			if f.ID != "" {
				if err := crawl(f.ID, fullP); err != nil {
					return err
				}
			}
		}
		return nil
	}
	startPath := rootPath
	if startPath == "" {
		startPath = "/"
	}
	if err := crawl(folderID, startPath); err != nil {
		return err
	}
	return saveIndex(rootPath, items)
}
