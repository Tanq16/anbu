package fssync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type ClientConfig struct {
	ServerAddr  string
	SyncDir     string
	IgnorePaths string
}

type Client struct {
	cfg          ClientConfig
	conn         *SafeConn
	watcher      *fsnotify.Watcher
	ignorer      *PathIgnorer
	syncingMutex sync.Mutex
	isSyncing    bool
}

func NewClient(cfg ClientConfig) (*Client, error) {
	if err := os.MkdirAll(cfg.SyncDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sync directory: %w", err)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}
	return &Client{
		cfg:     cfg,
		watcher: watcher,
		ignorer: NewPathIgnorer(cfg.IgnorePaths),
	}, nil
}

func (c *Client) Run(ctx context.Context) {
	defer c.watcher.Close()
	go c.watchFilesystem(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("Client shutting down...")
			if c.conn != nil && c.conn.Conn != nil {
				c.conn.Conn.Close()
			}
			return
		default:
		}
		err := c.connect()
		if err != nil {
			log.Error().Err(err).Msgf("Connection failed, retrying in 5 seconds...")
			select {
			case <-ctx.Done():
				log.Info().Msgf("Client shutting down...")
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}
		c.listenToServer(ctx)
		log.Warn().Msgf("Disconnected from server. Attempting to reconnect...")
		select {
		case <-ctx.Done():
			log.Info().Msgf("Client shutting down...")
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (c *Client) connect() error {
	u, err := url.Parse(c.cfg.ServerAddr)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}
	log.Info().Msgf("Connecting to server at %s...", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	c.conn = &SafeConn{Conn: conn}
	log.Info().Msgf("Successfully connected to server at %s", c.cfg.ServerAddr)
	return nil
}

func (c *Client) listenToServer(ctx context.Context) {
	defer c.conn.Conn.Close()
	go func() {
		<-ctx.Done()
		if c.conn != nil && c.conn.Conn != nil {
			c.conn.Conn.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("Stopping listener due to shutdown")
			return
		default:
		}
		var wrapper MessageWrapper
		if err := c.conn.Conn.ReadJSON(&wrapper); err != nil {
			// Don't log error if context was cancelled (expected shutdown)
			if ctx.Err() == nil {
				log.Error().Err(err).Msgf("Error reading from server")
			}
			return
		}
		switch wrapper.Type {
		case TypeManifest:
			go func() {
				c.setSyncing(true)
				c.handleManifest(wrapper.Payload)
				c.setSyncing(false)
			}()
		case TypeFileContent:
			c.setSyncing(true)
			c.handleFileContent(wrapper.Payload)
			c.setSyncing(false)
		case TypeFileOperation:
			c.setSyncing(true)
			c.handleFileOperation(wrapper.Payload)
			c.setSyncing(false)
		default:
			log.Warn().Msgf("Received unknown message type from server: %s", string(wrapper.Type))
		}
	}
}

func (c *Client) handleManifest(payload []byte) {
	var msg ManifestMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Error().Err(err).Msgf("Failed to unmarshal manifest")
		return
	}
	log.Info().Msgf("Received server manifest. Starting initial sync.")
	localManifest, err := BuildFileManifest(c.cfg.SyncDir, c.ignorer)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to build local manifest for sync")
		return
	}
	var toRequest []string
	serverFiles := make(map[string]bool)

	for path, serverHash := range msg.Files {
		serverFiles[path] = true
		localHash, exists := localManifest[path]
		if !exists || localHash != serverHash {
			toRequest = append(toRequest, path)
		}
	}

	for path := range localManifest {
		if !serverFiles[path] {
			fullPath := filepath.Join(c.cfg.SyncDir, path)
			log.Info().Msgf("Removing local file not present on server: %s", path)
			if err := os.RemoveAll(fullPath); err != nil {
				log.Error().Err(err).Msgf("Failed to remove local file: %s", fullPath)
			}
		}
	}
	if len(toRequest) > 0 {
		log.Info().Msgf("Requesting %d files from server", len(toRequest))
		c.requestFiles(toRequest)
	} else {
		log.Info().Msgf("Initial sync complete. Local directory is up-to-date.")
	}
}

func (c *Client) handleFileContent(payload []byte) {
	var msg FileContentMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Error().Err(err).Msgf("Failed to unmarshal file content")
		return
	}
	log.Info().Msgf("Received file content from server: %s", msg.Path)
	fullPath := filepath.Join(c.cfg.SyncDir, msg.Path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		log.Error().Err(err).Msgf("Failed to create parent directories: %s", fullPath)
		return
	}
	if err := os.WriteFile(fullPath, msg.Content, 0644); err != nil {
		log.Error().Err(err).Msgf("Failed to write file: %s", msg.Path)
	}
}

func (c *Client) handleFileOperation(payload []byte) {
	var op FileOperationMessage
	if err := json.Unmarshal(payload, &op); err != nil {
		log.Error().Err(err).Msgf("Failed to unmarshal file operation")
		return
	}
	log.Info().Msgf("Received file operation from server: op=%s path=%s", string(op.Op), op.Path)
	if err := ApplyOperation(c.cfg.SyncDir, &op); err != nil {
		log.Error().Err(err).Msgf("Failed to apply file operation: %s", op.Path)
	}
}

func (c *Client) requestFiles(paths []string) {
	payload, _ := json.Marshal(FileRequestMessage{Paths: paths})
	msg := MessageWrapper{
		Type:    TypeFileRequest,
		Payload: payload,
	}
	if err := c.conn.WriteJSON(msg); err != nil {
		log.Error().Err(err).Msgf("Failed to send file request to server")
	}
}

func (c *Client) watchFilesystem(ctx context.Context) {
	filepath.Walk(c.cfg.SyncDir, func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			relPath, _ := filepath.Rel(c.cfg.SyncDir, path)
			if !c.ignorer.IsIgnored(relPath) {
				if err := c.watcher.Add(path); err != nil {
					log.Error().Err(err).Msgf("Failed to add path to watcher: %s", path)
				}
			}
		}
		return nil
	})
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("Stopping filesystem watcher due to shutdown")
			return
		case event, ok := <-c.watcher.Events:
			if !ok {
				return
			}
			if c.isSyncing {
				continue
			}
			c.handleFsEvent(event)
		case err, ok := <-c.watcher.Errors:
			if !ok {
				return
			}
			log.Error().Err(err).Msgf("Watcher error")
		}
	}
}

func (c *Client) handleFsEvent(event fsnotify.Event) {
	relPath, err := filepath.Rel(c.cfg.SyncDir, event.Name)
	if err != nil || c.ignorer.IsIgnored(relPath) {
		return
	}
	op := FileOperationMessage{Path: relPath}
	if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
		op.Op = OpRemove
		c.watcher.Remove(event.Name)
	} else if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
		info, err := os.Stat(event.Name)
		if err != nil {
			return
		}
		op.Op = OpWrite
		if info.IsDir() {
			if event.Op&fsnotify.Create == fsnotify.Create {
				c.watcher.Add(event.Name)
				op.IsDir = true
			} else {
				return
			}
		} else {
			content, err := os.ReadFile(event.Name)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to read file for sending: %s", event.Name)
				return
			}
			op.Content = content
		}
	} else {
		return
	}
	log.Info().Msgf("Detected local change, sending to server: op=%s path=%s", string(op.Op), relPath)
	payload, _ := json.Marshal(op)
	msg := MessageWrapper{
		Type:    TypeFileOperation,
		Payload: payload,
	}
	if err := c.conn.WriteJSON(msg); err != nil {
		log.Error().Err(err).Msgf("Failed to send file operation to server")
	}
}

func (c *Client) setSyncing(status bool) {
	c.syncingMutex.Lock()
	defer c.syncingMutex.Unlock()
	c.isSyncing = status
}
