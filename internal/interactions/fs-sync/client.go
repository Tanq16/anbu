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
)

type ClientConfig struct {
	ServerAddr  string
	SyncDir     string
	IgnorePaths string
	DeleteExtra bool
	Insecure    bool
	DryRun      bool
}

type Client struct {
	cfg        ClientConfig
	ignorer    *PathIgnorer
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
		cfg:     cfg,
		ignorer: NewPathIgnorer(cfg.IgnorePaths),
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
	localManifest, _ := BuildManifest(c.cfg.SyncDir, c.ignorer)
	toRequest, toDelete := c.compareManifests(serverManifest, localManifest)
	if c.cfg.DryRun {
		log.Info().Msgf("Would sync %d files:", len(toRequest))
		for _, path := range toRequest {
			log.Info().Msgf("  - %s", path)
		}
		if c.cfg.DeleteExtra && len(toDelete) > 0 {
			log.Info().Msgf("Would delete %d files:", len(toDelete))
			for _, path := range toDelete {
				log.Info().Msgf("  - %s", path)
			}
		}
		return nil
	}
	if len(toRequest) > 0 {
		log.Info().Msgf("Syncing %d files...", len(toRequest))
		if err := c.fetchFiles(toRequest); err != nil {
			return fmt.Errorf("failed to fetch files: %w", err)
		}
	}
	if c.cfg.DeleteExtra && len(toDelete) > 0 {
		if err := c.deleteFiles(toDelete); err != nil {
			return fmt.Errorf("failed to delete files: %w", err)
		}
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

func (c *Client) fetchFiles(paths []string) error {
	reqBody, _ := json.Marshal(FileRequest{Paths: paths})
	resp, err := c.httpClient.Post(
		c.cfg.ServerAddr+"/files",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to fetch files: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var filesResp FilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&filesResp); err != nil {
		return err
	}
	for i, file := range filesResp.Files {
		fullPath := filepath.Join(c.cfg.SyncDir, file.Path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", file.Path, err)
		}
		if err := os.WriteFile(fullPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
		log.Info().Msgf("[%d/%d] %s", i+1, len(filesResp.Files), file.Path)
	}
	return nil
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

func (c *Client) deleteFiles(paths []string) error {
	for _, path := range paths {
		fullPath := filepath.Join(c.cfg.SyncDir, path)
		if err := os.RemoveAll(fullPath); err != nil {
			log.Warn().Err(err).Msgf("Failed to delete %s", path)
		} else {
			log.Info().Msgf("Deleted %s", path)
		}
	}
	return nil
}
