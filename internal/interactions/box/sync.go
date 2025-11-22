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

func buildLocalTree(rootDir string) (*FileTree, error) {
	tree := &FileTree{
		Files: make(map[string]FileInfo),
		Dirs:  make(map[string]*FileTree),
	}
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
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
				log.Error().Err(err).Msgf("Failed to compute hash for %s", path)
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

func buildRemoteTree(client *http.Client, folderID string, basePath string) (*FileTree, error) {
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
		itemPath := filepath.Join(basePath, item.Name)
		if item.Type == "folder" {
			subTree, err := buildRemoteTree(client, item.ID, itemPath)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to build tree for folder %s", itemPath)
				continue
			}
			tree.Dirs[item.Name] = subTree
		} else {
			hash, err := getRemoteFileHash(client, item.ID)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to get hash for file %s", itemPath)
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
		return handleBoxAPIError("upload file", resp)
	}
	return nil
}

func syncTree(client *http.Client, localTree *FileTree, remoteTree *FileTree, localBase string, remoteFolderID string) error {
	for fileName, localFile := range localTree.Files {
		remoteFile, exists := remoteTree.Files[fileName]
		localPath := filepath.Join(localBase, fileName)
		if !exists {
			fmt.Printf("Uploading %s\n", u.FDebug(localFile.Path))
			if err := uploadBoxFileToFolder(client, localPath, remoteFolderID); err != nil {
				log.Error().Err(err).Msgf("Failed to upload %s", localFile.Path)
			}
		} else if localFile.Hash != remoteFile.Hash {
			fmt.Printf("Updating %s\n", u.FDebug(localFile.Path))
			if err := deleteBoxFile(client, remoteFile.ID); err != nil {
				log.Error().Err(err).Msgf("Failed to delete %s for update", localFile.Path)
				continue
			}
			if err := uploadBoxFileToFolder(client, localPath, remoteFolderID); err != nil {
				log.Error().Err(err).Msgf("Failed to upload updated %s", localFile.Path)
			}
		}
	}
	for fileName, remoteFile := range remoteTree.Files {
		if _, exists := localTree.Files[fileName]; !exists {
			fmt.Printf("%s %s\n", u.FError("Deleting"), u.FDebug(remoteFile.Path))
			if err := deleteBoxFile(client, remoteFile.ID); err != nil {
				log.Error().Err(err).Msgf("Failed to delete %s", remoteFile.Path)
			}
		}
	}
	for dirName, localSubTree := range localTree.Dirs {
		remoteSubTree, exists := remoteTree.Dirs[dirName]
		var subFolderID string
		if !exists {
			folderJSON := fmt.Sprintf(`{"name":"%s", "parent":{"id":"%s"}}`, dirName, remoteFolderID)
			req, err := http.NewRequest("POST", uploadFolderURL, bytes.NewBufferString(folderJSON))
			if err != nil {
				log.Error().Err(err).Msgf("Failed to create folder request for %s", dirName)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to create folder %s", dirName)
				continue
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				log.Error().Msgf("Failed to create folder %s (status %d)", dirName, resp.StatusCode)
				continue
			}
			var folder BoxItem
			if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
				log.Error().Err(err).Msgf("Failed to parse folder response for %s", dirName)
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
				log.Error().Msgf("Failed to find folder ID for %s", dirName)
				continue
			}
		}
		subLocalBase := filepath.Join(localBase, dirName)
		if err := syncTree(client, localSubTree, remoteSubTree, subLocalBase, subFolderID); err != nil {
			log.Error().Err(err).Msgf("Failed to sync subtree %s", dirName)
		}
	}
	for dirName := range remoteTree.Dirs {
		if _, exists := localTree.Dirs[dirName]; !exists {
			req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, remoteFolderID), nil)
			if err != nil {
				continue
			}
			q := req.URL.Query()
			q.Add("fields", "type,name,id")
			req.URL.RawQuery = q.Encode()
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				continue
			}
			var items BoxFolderItems
			if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
				continue
			}
			for _, item := range items.Entries {
				if item.Type == "folder" && item.Name == dirName {
					if err := deleteBoxFolderRecursive(client, item.ID); err != nil {
						log.Error().Err(err).Msgf("Failed to delete folder %s", dirName)
					} else {
						fmt.Printf("%s %s\n", u.FError("Deleting folder"), u.FDebug(dirName))
					}
					break
				}
			}
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

func SyncBoxDirectory(client *http.Client, localDir string, remotePath string) error {
	localTree, err := buildLocalTree(localDir)
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
	remoteTree, err := buildRemoteTree(client, folderID, "")
	if err != nil {
		return fmt.Errorf("failed to build remote tree: %v", err)
	}
	return syncTree(client, localTree, remoteTree, localDir, folderID)
}
