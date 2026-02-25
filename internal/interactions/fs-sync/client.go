package fssync

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	u "github.com/tanq16/anbu/internal/utils"
)

type ClientConfig struct {
	ServerAddr  string
	SyncDir     string
	DeleteExtra bool
	Insecure    bool
	DryRun      bool
	IgnorePaths string
}

type Client struct {
	cfg        ClientConfig
	httpClient *http.Client
	ignorer    *PathIgnorer
}

func NewClient(cfg ClientConfig) (*Client, error) {
	absDir, err := filepath.Abs(cfg.SyncDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve sync directory: %w", err)
	}
	cfg.SyncDir = absDir
	if err := os.MkdirAll(cfg.SyncDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sync directory: %w", err)
	}
	transport := &http.Transport{}
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout:   5 * time.Minute,
			Transport: transport,
		},
		ignorer: NewPathIgnorer(cfg.IgnorePaths),
	}, nil
}

func (c *Client) Run() error {
	mode, err := c.fetchMode()
	if err != nil {
		return fmt.Errorf("failed to detect server mode: %w", err)
	}
	u.PrintInfo(fmt.Sprintf("Server mode: %s", mode))
	switch mode {
	case "send":
		return c.pullFromServer()
	case "receive":
		return c.pushToServer()
	default:
		return fmt.Errorf("unknown server mode: %s", mode)
	}
}

func (c *Client) fetchMode() (string, error) {
	resp, err := c.httpClient.Get(c.cfg.ServerAddr + "/mode")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned %d", resp.StatusCode)
	}
	var modeResp ModeResponse
	if err := json.NewDecoder(resp.Body).Decode(&modeResp); err != nil {
		return "", err
	}
	return modeResp.Mode, nil
}

// pullFromServer handles the case where the server is sending files to us
func (c *Client) pullFromServer() error {
	serverManifest, err := c.fetchManifest()
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}

	localManifest, _ := BuildManifest(c.cfg.SyncDir, c.ignorer)

	filteredServer := make(map[string]string, len(serverManifest))
	for path, hash := range serverManifest {
		if !c.ignorer.IsIgnored(path) {
			filteredServer[path] = hash
		}
	}

	toRequest, toDelete := c.compareManifests(filteredServer, localManifest)
	if c.cfg.DryRun {
		for _, path := range toRequest {
			u.PrintGeneric(fmt.Sprintf("Dry Run: %s", u.FDebug(path)))
		}
		if c.cfg.DeleteExtra {
			for _, path := range toDelete {
				u.PrintGeneric(fmt.Sprintf("Dry Run (delete): %s", u.FDebug(path)))
			}
		}
		u.LineBreak()
		totalCount := len(toRequest)
		if c.cfg.DeleteExtra {
			totalCount += len(toDelete)
		}
		if totalCount == 0 {
			u.PrintWarn("no files would be synced", nil)
		} else {
			u.PrintGeneric(fmt.Sprintf("%s %s", u.FDebug("Operation completed:"), u.FSuccess(fmt.Sprintf("%d file(s) would be synced", totalCount))))
		}
		return nil
	}

	syncedCount := 0
	if len(toRequest) > 0 {
		syncedCount, err = c.fetchFiles(toRequest)
		if err != nil {
			return fmt.Errorf("failed to fetch files: %w", err)
		}
	}
	deletedCount := 0
	if c.cfg.DeleteExtra && len(toDelete) > 0 {
		deletedCount, err = c.deleteLocalFiles(toDelete)
		if err != nil {
			return fmt.Errorf("failed to delete files: %w", err)
		}
	}
	u.LineBreak()
	totalCount := syncedCount + deletedCount
	if totalCount == 0 {
		u.PrintWarn("no files were synced", nil)
	} else {
		u.PrintGeneric(fmt.Sprintf("%s %s", u.FDebug("Operation completed:"), u.FSuccess(fmt.Sprintf("%d file(s) synced", totalCount))))
	}
	return nil
}

// pushToServer handles the case where the server is receiving files from us
func (c *Client) pushToServer() error {
	serverManifest, err := c.fetchManifest()
	if err != nil {
		return fmt.Errorf("failed to fetch server manifest: %w", err)
	}

	localManifest, err := BuildManifest(c.cfg.SyncDir, c.ignorer)
	if err != nil {
		return fmt.Errorf("failed to build local manifest: %w", err)
	}

	var needed []string
	for path, localHash := range localManifest {
		if serverHash, exists := serverManifest[path]; !exists || serverHash != localHash {
			needed = append(needed, path)
		}
	}

	var toDelete []string
	for path := range serverManifest {
		if _, exists := localManifest[path]; !exists {
			toDelete = append(toDelete, path)
		}
	}

	if len(needed) == 0 && len(toDelete) == 0 {
		u.PrintWarn("no files to send", nil)
		return nil
	}

	var files []FileContent
	for _, path := range needed {
		fullPath := filepath.Join(c.cfg.SyncDir, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			u.PrintWarn(fmt.Sprintf("Failed to read %s", path), err)
			continue
		}
		files = append(files, FileContent{Path: path, Content: content})
	}

	uploadReq := UploadRequest{
		Files:    files,
		ToDelete: toDelete,
	}
	reqBody, _ := json.Marshal(uploadReq)
	resp, err := c.httpClient.Post(
		c.cfg.ServerAddr+"/upload",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to upload files: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}

	u.LineBreak()
	for _, file := range files {
		u.PrintGeneric(fmt.Sprintf("Sent: %s", u.FSuccess(file.Path)))
	}
	u.PrintGeneric(fmt.Sprintf("%s %s", u.FDebug("Operation completed:"), u.FSuccess(fmt.Sprintf("%d file(s) sent", len(files)))))
	return nil
}

func (c *Client) fetchManifest() (map[string]string, error) {
	resp, err := c.httpClient.Get(c.cfg.ServerAddr + "/manifest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}
	var manifest ManifestResponse
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}
	return manifest.Files, nil
}

func (c *Client) fetchFiles(paths []string) (int, error) {
	reqBody, _ := json.Marshal(FileRequest{Paths: paths})
	resp, err := c.httpClient.Post(
		c.cfg.ServerAddr+"/files",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var filesResp FilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&filesResp); err != nil {
		return 0, err
	}
	count := 0
	for _, file := range filesResp.Files {
		fullPath := filepath.Join(c.cfg.SyncDir, file.Path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return count, fmt.Errorf("failed to create directory for %s: %w", file.Path, err)
		}
		if err := os.WriteFile(fullPath, file.Content, 0644); err != nil {
			return count, fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
		u.PrintGeneric(fmt.Sprintf("Synced: %s", u.FSuccess(file.Path)))
		count++
	}
	return count, nil
}

func (c *Client) compareManifests(server, local map[string]string) (toRequest, toDelete []string) {
	for path, serverHash := range server {
		if localHash, exists := local[path]; !exists || localHash != serverHash {
			toRequest = append(toRequest, path)
		}
	}
	for path := range local {
		if _, exists := server[path]; !exists {
			toDelete = append(toDelete, path)
		}
	}
	return
}

func (c *Client) deleteLocalFiles(paths []string) (int, error) {
	count := 0
	for _, path := range paths {
		fullPath := filepath.Join(c.cfg.SyncDir, path)
		if err := os.RemoveAll(fullPath); err != nil {
			u.PrintError(fmt.Sprintf("Failed to delete %s", path), err)
		} else {
			u.PrintGeneric(fmt.Sprintf("Deleted: %s", u.FSuccess(path)))
			count++
		}
	}
	return count, nil
}
