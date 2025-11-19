package box

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	boxTokenFile    = ".anbu-box-token.json"
	redirectURI     = "http://localhost:8080"
	apiBaseURL      = "https://api.box.com/2.0"
	uploadBaseURL   = "https://upload.box.com/api/2.0"
	folderItemsURL  = apiBaseURL + "/folders/%s/items"
	fileContentURL  = apiBaseURL + "/files/%s/content"
	uploadFileURL   = uploadBaseURL + "/files/content"
	uploadFolderURL = apiBaseURL + "/folders"
)

type BoxCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type BoxItem struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Size   *int64 `json:"size"`
	Parent *struct {
		ID string `json:"id"`
	} `json:"parent"`
	ModifiedAt *string `json:"modified_at"`
}

type BoxFolderItems struct {
	TotalCount int       `json:"total_count"`
	Entries    []BoxItem `json:"entries"`
}

type BoxError struct {
	Type    string `json:"type"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BoxItemDisplay struct {
	Name         string
	ModifiedTime string
	Size         int64
	Type         string
}

func ResolvePath(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path, nil
	}
	shortcutsFile := filepath.Join(homeDir, ".anbu-box-shortcuts.json")
	data, err := os.ReadFile(shortcutsFile)
	if err != nil {
		return path, nil
	}
	var shortcuts map[string]string
	if err := json.Unmarshal(data, &shortcuts); err != nil {
		return path, nil
	}
	result := strings.Builder{}
	i := 0
	for i < len(path) {
		if path[i] == '%' {
			if i+1 < len(path) && path[i+1] == '%' {
				result.WriteByte('%')
				i += 2
				continue
			}
			j := i + 1
			for j < len(path) && (path[j] >= 'a' && path[j] <= 'z' || path[j] >= 'A' && path[j] <= 'Z' || path[j] >= '0' && path[j] <= '9' || path[j] == '_') {
				j++
			}
			if j > i+1 {
				key := path[i+1 : j]
				if val, ok := shortcuts[key]; ok {
					result.WriteString(val)
					i = j
					continue
				}
			}
			result.WriteByte('%')
			i++
		} else {
			result.WriteByte(path[i])
			i++
		}
	}
	return result.String(), nil
}
