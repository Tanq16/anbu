package gdrive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
)

const (
	gdriveTokenFile      = ".anbu-gdrive-token.json"
	googleFolderMimeType = "application/vnd.google-apps.folder"
)

type DriveItem struct {
	Name         string
	ModifiedTime string
	Size         int64
}

func ResolvePath(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path, nil
	}
	shortcutsFile := filepath.Join(homeDir, ".anbu-gdrive-shortcuts.json")
	data, err := os.ReadFile(shortcutsFile)
	if err != nil {
		return path, nil
	}
	var shortcuts map[string]string
	if err := json.Unmarshal(data, &shortcuts); err != nil {
		return path, nil
	}
	result := path
	result = regexp.MustCompile(`%%`).ReplaceAllString(result, "%")
	pattern := regexp.MustCompile(`%([^%]+)%`)
	result = pattern.ReplaceAllStringFunc(result, func(match string) string {
		key := match[1 : len(match)-1]
		if val, ok := shortcuts[key]; ok {
			return val
		}
		return match
	})
	return result, nil
}
