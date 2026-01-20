package box

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
)

type FileTree struct {
	Files map[string]FileInfo
	Dirs  map[string]*FileTree
}

type FileInfo struct {
	Path string
	Hash string
	ID   string
}

func computeLocalHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func buildLocalTree(rootDir string, ignoreSet map[string]struct{}) (*FileTree, error) {
	tree := &FileTree{
		Files: make(map[string]FileInfo),
		Dirs:  make(map[string]*FileTree),
	}
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if _, skip := ignoreSet[d.Name()]; skip {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		if d.IsDir() {
			parts := strings.Split(relPath, string(filepath.Separator))
			current := tree
			for _, part := range parts {
				if current.Dirs == nil {
					current.Dirs = make(map[string]*FileTree)
				}
				if _, exists := current.Dirs[part]; !exists {
					current.Dirs[part] = &FileTree{
						Files: make(map[string]FileInfo),
						Dirs:  make(map[string]*FileTree),
					}
				}
				current = current.Dirs[part]
			}
		} else {
			hash, err := computeLocalHash(path)
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to compute hash for %s", path), err)
				return nil
			}
			parts := strings.Split(relPath, string(filepath.Separator))
			fileName := parts[len(parts)-1]
			dirPath := strings.Join(parts[:len(parts)-1], string(filepath.Separator))
			current := tree
			if dirPath != "" {
				for _, part := range strings.Split(dirPath, string(filepath.Separator)) {
					if current.Dirs == nil {
						current.Dirs = make(map[string]*FileTree)
					}
					if _, exists := current.Dirs[part]; !exists {
						current.Dirs[part] = &FileTree{
							Files: make(map[string]FileInfo),
							Dirs:  make(map[string]*FileTree),
						}
					}
					current = current.Dirs[part]
				}
			}
			current.Files[fileName] = FileInfo{
				Path: relPath,
				Hash: hash,
			}
		}
		return nil
	})
	return tree, err
}

func getRemoteFileHash(client *http.Client, fileID string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/files/%s", apiBaseURL, fileID), nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("fields", "sha1")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get file hash: status %d", resp.StatusCode)
	}
	var file struct {
		SHA1 string `json:"sha1"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return "", err
	}
	return file.SHA1, nil
}

func buildRemoteTree(client *http.Client, folderID string, basePath string, ignoreSet map[string]struct{}) (*FileTree, error) {
	tree := &FileTree{
		Files: make(map[string]FileInfo),
		Dirs:  make(map[string]*FileTree),
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, folderID), nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("fields", "type,name,id")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, handleBoxAPIError("list folder", resp)
	}
	var items BoxFolderItems
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}
	for _, item := range items.Entries {
		if _, skip := ignoreSet[item.Name]; skip {
			continue
		}
		itemPath := filepath.Join(basePath, item.Name)
		if item.Type == "folder" {
			subTree, err := buildRemoteTree(client, item.ID, itemPath, ignoreSet)
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to build tree for folder %s", itemPath), err)
				continue
			}
			tree.Dirs[item.Name] = subTree
		} else {
			hash, err := getRemoteFileHash(client, item.ID)
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to get hash for file %s", itemPath), err)
				continue
			}
			tree.Files[item.Name] = FileInfo{
				Path: itemPath,
				Hash: hash,
				ID:   item.ID,
			}
		}
	}
	return tree, nil
}

func deleteBoxFile(client *http.Client, fileID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/files/%s", apiBaseURL, fileID), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return handleBoxAPIError("delete file", resp)
	}
	return nil
}

func findFileIDInFolder(client *http.Client, fileName string, folderID string) (string, error) {
	offset := 0
	limit := 1000
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, folderID), nil)
		if err != nil {
			return "", err
		}
		q := req.URL.Query()
		q.Add("fields", "type,name,id")
		q.Add("limit", fmt.Sprintf("%d", limit))
		q.Add("offset", fmt.Sprintf("%d", offset))
		req.URL.RawQuery = q.Encode()
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return "", fmt.Errorf("failed to list folder items (status %d)", resp.StatusCode)
		}
		var items BoxFolderItems
		if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
			resp.Body.Close()
			return "", err
		}
		resp.Body.Close()
		for _, item := range items.Entries {
			if item.Type == "file" && item.Name == fileName {
				return item.ID, nil
			}
		}
		if len(items.Entries) < limit {
			break
		}
		offset += limit
	}
	return "", fmt.Errorf("file not found in folder after checking all pages")
}

func uploadBoxFileVersion(client *http.Client, localPath string, fileID string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileName := filepath.Base(localPath)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	writer.Close()
	req, err := http.NewRequest("POST", fmt.Sprintf(uploadFileVersionURL, fileID), body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("If-Match", "*")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return handleBoxAPIError("upload file version", resp)
	}
	return nil
}

func uploadBoxFileToFolder(client *http.Client, localPath string, parentFolderID string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileName := filepath.Base(localPath)
	attributesJSON := fmt.Sprintf(`{"name":"%s", "parent":{"id":"%s"}}`, fileName, parentFolderID)
	if err := writer.WriteField("attributes", attributesJSON); err != nil {
		return err
	}
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	writer.Close()
	req, err := http.NewRequest("POST", uploadFileURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr == nil {
			var boxErr BoxError
			if json.Unmarshal(bodyBytes, &boxErr) == nil {
				if boxErr.Code == "item_name_in_use" {
					log.Debug().Str("file", fileName).Str("folder", parentFolderID).Msg("File already exists, attempting to update version")
					fileID, findErr := findFileIDInFolder(client, fileName, parentFolderID)
					if findErr == nil {
						return uploadBoxFileVersion(client, localPath, fileID)
					}
					return fmt.Errorf("file '%s' already exists but couldn't find its ID to update: %v", fileName, findErr)
				}
				return fmt.Errorf("api request to 'upload file' failed: %s - %s", boxErr.Code, boxErr.Message)
			}
			return fmt.Errorf("api request to 'upload file' failed with status %s: %s", resp.Status, string(bodyBytes))
		}
		return handleBoxAPIError("upload file", resp)
	}
	return nil
}

func syncTree(client *http.Client, localTree *FileTree, remoteTree *FileTree, localBase string, remoteFolderID string, sem chan struct{}, wg *sync.WaitGroup, depth int) error {
	for fileName, localFile := range localTree.Files {
		remoteFile, exists := remoteTree.Files[fileName]
		localPath := filepath.Join(localBase, fileName)
		switch {
		case !exists:
			sem <- struct{}{}
			wg.Add(1)
			go func(localPath string, localFile FileInfo) {
				defer wg.Done()
				defer func() { <-sem }()
				u.PrintStream(fmt.Sprintf("Uploading %s", u.FDebug(localFile.Path)))
				if err := uploadBoxFileToFolder(client, localPath, remoteFolderID); err != nil {
					u.PrintError(fmt.Sprintf("Failed to upload %s", localFile.Path), err)
				}
			}(localPath, localFile)
		case localFile.Hash != remoteFile.Hash:
			sem <- struct{}{}
			wg.Add(1)
			go func(localPath string, localFile FileInfo, remoteFile FileInfo) {
				defer wg.Done()
				defer func() { <-sem }()
				log.Debug().Str("path", localFile.Path).Msg("file hash mismatch, update needed")
				u.PrintStream(fmt.Sprintf("Updating %s", u.FDebug(localFile.Path)))
				if err := deleteBoxFile(client, remoteFile.ID); err != nil {
					u.PrintError(fmt.Sprintf("Failed to delete %s for update", localFile.Path), err)
					return
				}
				if err := uploadBoxFileToFolder(client, localPath, remoteFolderID); err != nil {
					u.PrintError(fmt.Sprintf("Failed to upload updated %s", localFile.Path), err)
				}
			}(localPath, localFile, remoteFile)
		}
	}
	for fileName, remoteFile := range remoteTree.Files {
		if _, exists := localTree.Files[fileName]; !exists {
			sem <- struct{}{}
			wg.Add(1)
			go func(remoteFile FileInfo) {
				defer wg.Done()
				defer func() { <-sem }()
				u.PrintGeneric(fmt.Sprintf("%s %s", u.FError("Deleting"), u.FDebug(remoteFile.Path)))
				if err := deleteBoxFile(client, remoteFile.ID); err != nil {
					u.PrintError(fmt.Sprintf("Failed to delete %s", remoteFile.Path), err)
				}
			}(remoteFile)
		}
	}
	for dirName, localSubTree := range localTree.Dirs {
		if depth == 0 {
			log.Debug().Str("directory", dirName).Msg("Syncing directory")
		}
		remoteSubTree, exists := remoteTree.Dirs[dirName]
		var subFolderID string
		if !exists {
			log.Debug().Str("dir", dirName).Msg("folder not found in remote tree, creating new folder")
			folderJSON := fmt.Sprintf(`{"name":"%s", "parent":{"id":"%s"}}`, dirName, remoteFolderID)
			req, err := http.NewRequest("POST", uploadFolderURL, bytes.NewBufferString(folderJSON))
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to create folder request for %s", dirName), err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to create folder %s", dirName), err)
				continue
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				u.PrintError(fmt.Sprintf("Failed to create folder %s (status %d)", dirName, resp.StatusCode), nil)
				continue
			}
			var folder BoxItem
			if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
				u.PrintError(fmt.Sprintf("Failed to parse folder response for %s", dirName), err)
				continue
			}
			subFolderID = folder.ID
			remoteSubTree = &FileTree{
				Files: make(map[string]FileInfo),
				Dirs:  make(map[string]*FileTree),
			}
		} else {
			req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, remoteFolderID), nil)
			if err == nil {
				q := req.URL.Query()
				q.Add("fields", "type,name,id")
				req.URL.RawQuery = q.Encode()
				resp, err := client.Do(req)
				if err == nil && resp.StatusCode == http.StatusOK {
					var items BoxFolderItems
					if json.NewDecoder(resp.Body).Decode(&items) == nil {
						for _, item := range items.Entries {
							if item.Type == "folder" && item.Name == dirName {
								subFolderID = item.ID
								break
							}
						}
					}
					resp.Body.Close()
				}
			}
			if subFolderID == "" {
				u.PrintError(fmt.Sprintf("Failed to find folder ID for %s", dirName), nil)
				continue
			} else {
				log.Debug().Str("dir", dirName).Str("folderID", subFolderID).Msg("folder found in remote tree")
			}
		}
		subLocalBase := filepath.Join(localBase, dirName)
		if err := syncTree(client, localSubTree, remoteSubTree, subLocalBase, subFolderID, sem, wg, depth+1); err != nil {
			u.PrintError(fmt.Sprintf("Failed to sync subtree %s", dirName), err)
		}
	}
	for dirName := range remoteTree.Dirs {
		if _, exists := localTree.Dirs[dirName]; !exists {
			sem <- struct{}{}
			wg.Add(1)
			go func(dirName string) {
				defer wg.Done()
				defer func() { <-sem }()
				req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, remoteFolderID), nil)
				if err != nil {
					return
				}
				q := req.URL.Query()
				q.Add("fields", "type,name,id")
				req.URL.RawQuery = q.Encode()
				resp, err := client.Do(req)
				if err != nil {
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					return
				}
				var items BoxFolderItems
				if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
					return
				}
				for _, item := range items.Entries {
					if item.Type == "folder" && item.Name == dirName {
						if err := deleteBoxFolderRecursive(client, item.ID); err != nil {
							u.PrintError(fmt.Sprintf("Failed to delete folder %s", dirName), err)
						} else {
							u.PrintGeneric(fmt.Sprintf("%s %s", u.FError("Deleting folder"), u.FDebug(dirName)))
						}
						break
					}
				}
			}(dirName)
		}
	}
	return nil
}

func deleteBoxFolderRecursive(client *http.Client, folderID string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, folderID), nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("fields", "type,id")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return handleBoxAPIError("list folder for delete", resp)
	}
	var items BoxFolderItems
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return err
	}
	for _, item := range items.Entries {
		if item.Type == "folder" {
			if err := deleteBoxFolderRecursive(client, item.ID); err != nil {
				return err
			}
		} else {
			if err := deleteBoxFile(client, item.ID); err != nil {
				return err
			}
		}
	}
	req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/folders/%s?recursive=true", apiBaseURL, folderID), nil)
	if err != nil {
		return err
	}
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return handleBoxAPIError("delete folder", resp)
	}
	return nil
}

func SyncBoxDirectory(client *http.Client, localDir string, remotePath string, concurrency int, ignore []string) error {
	ignoreSet := make(map[string]struct{})
	for _, v := range ignore {
		name := strings.TrimSpace(v)
		if name != "" {
			ignoreSet[name] = struct{}{}
		}
	}
	log.Debug().Str("localDir", localDir).Msg("building local tree")
	localTree, err := buildLocalTree(localDir, ignoreSet)
	if err != nil {
		return fmt.Errorf("failed to build local tree: %v", err)
	}
	folderID := "0"
	if remotePath != "" && remotePath != "/" && remotePath != "root" {
		var err error
		folderID, _, err = resolvePathToID(client, remotePath, "folder")
		if err != nil {
			return fmt.Errorf("failed to resolve remote path: %v", err)
		}
	}
	log.Debug().Str("remotePath", remotePath).Str("folderID", folderID).Msg("building remote tree")
	remoteTree, err := buildRemoteTree(client, folderID, "", ignoreSet)
	if err != nil {
		return fmt.Errorf("failed to build remote tree: %v", err)
	}
	if concurrency < 1 {
		concurrency = 1
	}

	totalTopLevelDirs := len(localTree.Dirs)
	totalTopLevelFiles := len(localTree.Files)
	if totalTopLevelDirs > 0 || totalTopLevelFiles > 0 {
		log.Debug().Int("directories", totalTopLevelDirs).Int("files", totalTopLevelFiles).Msg("Starting sync")
	}
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	err = syncTree(client, localTree, remoteTree, localDir, folderID, sem, &wg, 0)
	wg.Wait()
	if err == nil {
		log.Debug().Msg("Sync completed")
	}
	return err
}
