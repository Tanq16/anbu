package fssync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type ServerConfig struct {
	Port        int
	SyncDir     string
	IgnorePaths string
}

type clientConnection struct {
	id   string
	conn *SafeConn
}

type fileOperationEnvelope struct {
	senderID string
	op       FileOperationMessage
}

type Server struct {
	cfg       ServerConfig
	clients   sync.Map
	ignorer   *PathIgnorer
	opChan    chan fileOperationEnvelope
	diskMutex sync.Mutex
}

func NewServer(cfg ServerConfig) (*Server, error) {
	if err := os.MkdirAll(cfg.SyncDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sync directory: %w", err)
	}
	return &Server{
		cfg:     cfg,
		ignorer: NewPathIgnorer(cfg.IgnorePaths),
		opChan:  make(chan fileOperationEnvelope, 100),
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	go s.processOperationQueue(ctx)
	http.HandleFunc("/ws", s.handleConnections)
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	server := &http.Server{
		Addr: addr,
	}
	serverErr := make(chan error, 1)
	go func() {
		log.Info().Msgf("WebSocket server starting to listen on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()
	<-ctx.Done()
	log.Info().Msgf("Shutting down server...")
	shutdownCtx := context.Background()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msgf("Error during server shutdown")
		return err
	}
	select {
	case err := <-serverErr:
		return err
	default:
		return nil
	}
}

func (s *Server) processOperationQueue(ctx context.Context) {
	log.Info().Msgf("Starting file operation queue processor")
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("Operation queue processor shutting down")
			return
		case envelope, ok := <-s.opChan:
			if !ok {
				return
			}
			log.Info().Msgf("Processing operation from queue: op=%s path=%s client_id=%s", string(envelope.op.Op), envelope.op.Path, envelope.senderID)
			s.applyChangeLocally(&envelope.op)
			s.broadcastOperation(envelope.senderID, &envelope.op)
		}
	}
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to upgrade connection")
		return
	}
	defer ws.Close()
	client := &clientConnection{
		id:   uuid.NewString(),
		conn: &SafeConn{Conn: ws},
	}
	s.clients.Store(client.id, client)
	log.Info().Msgf("Client connected: client_id=%s addr=%s", client.id, ws.RemoteAddr().String())
	defer func() {
		s.clients.Delete(client.id)
		log.Info().Msgf("Client disconnected: client_id=%s", client.id)
	}()
	if err := s.sendInitialManifest(client); err != nil {
		log.Error().Err(err).Msgf("Failed to send initial manifest: client_id=%s", client.id)
		return
	}
	s.handleClientMessages(client)
}

func (s *Server) sendInitialManifest(client *clientConnection) error {
	log.Info().Msgf("Building and sending initial manifest: client_id=%s", client.id)
	manifest, err := BuildFileManifest(s.cfg.SyncDir, s.ignorer)
	if err != nil {
		return fmt.Errorf("could not build file manifest: %w", err)
	}
	payload, _ := json.Marshal(ManifestMessage{Files: manifest})
	msg := MessageWrapper{
		Type:    TypeManifest,
		Payload: payload,
	}
	return client.conn.WriteJSON(msg)
}

func (s *Server) handleClientMessages(client *clientConnection) {
	for {
		var wrapper MessageWrapper
		if err := client.conn.Conn.ReadJSON(&wrapper); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msgf("Client read error: client_id=%s", client.id)
			}
			break
		}
		switch wrapper.Type {
		case TypeFileRequest:
			s.handleFileRequest(client, wrapper.Payload)
		case TypeFileOperation:
			s.handleFileOperation(client, wrapper.Payload)
		default:
			log.Warn().Msgf("Received unknown message type from client: %s", string(wrapper.Type))
		}
	}
}

func (s *Server) handleFileRequest(client *clientConnection, payload []byte) {
	var req FileRequestMessage
	if err := json.Unmarshal(payload, &req); err != nil {
		log.Error().Err(err).Msgf("Failed to unmarshal file request")
		return
	}
	log.Info().Msgf("Handling file request: count=%d client_id=%s", len(req.Paths), client.id)
	for _, path := range req.Paths {
		if s.ignorer.IsIgnored(path) {
			continue
		}
		fullPath := filepath.Join(s.cfg.SyncDir, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to read file for client request: %s", path)
			continue
		}
		contentPayload, _ := json.Marshal(FileContentMessage{Path: path, Content: content})
		msg := MessageWrapper{
			Type:    TypeFileContent,
			Payload: contentPayload,
		}
		if err := client.conn.WriteJSON(msg); err != nil {
			log.Error().Err(err).Msgf("Failed to send file content: client_id=%s", client.id)
			break
		}
	}
}

func (s *Server) handleFileOperation(sender *clientConnection, payload []byte) {
	var op FileOperationMessage
	if err := json.Unmarshal(payload, &op); err != nil {
		log.Error().Err(err).Msgf("Failed to unmarshal file operation")
		return
	}
	if s.ignorer.IsIgnored(op.Path) {
		log.Debug().Msgf("Ignoring file operation based on server rules: %s", op.Path)
		return
	}
	log.Debug().Msgf("Received and queuing file operation: path=%s client_id=%s", op.Path, sender.id)
	s.opChan <- fileOperationEnvelope{
		senderID: sender.id,
		op:       op,
	}
}

func (s *Server) applyChangeLocally(op *FileOperationMessage) {
	s.diskMutex.Lock()
	defer s.diskMutex.Unlock()
	if err := ApplyOperation(s.cfg.SyncDir, op); err != nil {
		log.Error().Err(err).Msgf("Failed to apply operation locally: %s", op.Path)
	}
}

func (s *Server) broadcastOperation(senderID string, op *FileOperationMessage) {
	payload, _ := json.Marshal(op)
	msg := MessageWrapper{
		Type:    TypeFileOperation,
		Payload: payload,
	}
	s.clients.Range(func(key, value any) bool {
		id := key.(string)
		client := value.(*clientConnection)
		if id != senderID {
			if err := client.conn.WriteJSON(msg); err != nil {
				log.Error().Err(err).Msgf("Failed to broadcast operation: client_id=%s", id)
			}
		}
		return true
	})
}
