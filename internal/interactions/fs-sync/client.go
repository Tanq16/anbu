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

	"github.com/rs/zerolog/log"

	u "github.com/tanq16/anbu/utils"
)

type ClientConfig struct {
	ServerAddr  string
	SyncDir     string
	DeleteExtra bool
	Insecure    bool
	DryRun      bool
}

type Client struct {
	cfg        ClientConfig
	httpClient *http.Client
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
	}, nil
}

func (c *Client) Run() error {
	serverManifest, err := c.fetchManifest()
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}
	localManifest, _ := BuildManifest(c.cfg.SyncDir, nil)
	toRequest, toDelete := c.compareManifests(serverManifest, localManifest)
	if c.cfg.DryRun {
		for _, path := range toRequest {
			fmt.Printf("Dry Run: %s\n", u.FDebug(path))
		}
		if c.cfg.DeleteExtra {
			for _, path := range toDelete {
				fmt.Printf("Dry Run: %s\n", u.FDebug(path))
			}
		}
		fmt.Println()
		totalCount := len(toRequest)
		if c.cfg.DeleteExtra {
			totalCount += len(toDelete)
		}
		if totalCount == 0 {
			log.Warn().Msg("no files would be synced")
		} else {
			fmt.Printf("%s %s\n", u.FDebug("Operation completed:"),
				u.FSuccess(fmt.Sprintf("%d file(s) would be synced", totalCount)))
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
		deletedCount, err = c.deleteFiles(toDelete)
		if err != nil {
			return fmt.Errorf("failed to delete files: %w", err)
		}
	}
	fmt.Println()
	totalCount := syncedCount + deletedCount
	if totalCount == 0 {
		log.Warn().Msg("no files were synced")
	} else {
		fmt.Printf("%s %s\n", u.FDebug("Operation completed:"),
			u.FSuccess(fmt.Sprintf("%d file(s) synced", totalCount)))
	}
	return nil
}

func (c *Client) fetchManifest() (map[string]string, error) {
	resp, err := c.httpClient.Get(c.cfg.ServerAddr + "/manifest")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
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
		return 0, fmt.Errorf("failed to fetch files: %w", err)
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
		fmt.Printf("Synced: %s\n", u.FSuccess(file.Path))
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

func (c *Client) deleteFiles(paths []string) (int, error) {
	count := 0
	for _, path := range paths {
		fullPath := filepath.Join(c.cfg.SyncDir, path)
		if err := os.RemoveAll(fullPath); err != nil {
			u.PrintError(fmt.Sprintf("Failed to delete %s: %v", path, err))
		} else {
			fmt.Printf("Deleted: %s\n", u.FSuccess(path))
			count++
		}
	}
	return count, nil
}
