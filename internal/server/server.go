package server

import (
	"embed"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/tanq16/anbu/internal/awscred"
	"github.com/tanq16/anbu/internal/awsrun"
	"github.com/tanq16/anbu/internal/datadir"
	"github.com/tanq16/anbu/internal/hosts"
	"github.com/tanq16/anbu/internal/jobs"
	"github.com/tanq16/anbu/internal/machine"
	"github.com/tanq16/anbu/internal/scaffold"
	"github.com/tanq16/anbu/internal/settings"
	"github.com/tanq16/anbu/internal/sshx"
	"github.com/tanq16/anbu/internal/tasks"
	"github.com/tanq16/anbu/internal/vault"
)

//go:embed static
var staticFiles embed.FS

const maxBody = 4 << 20

var errBadRequest = errors.New("bad request")

type Config struct {
	Host    string
	Port    int
	DataDir string
}

type Server struct {
	cfg        Config
	mux        *http.ServeMux
	vault      *vault.Store
	tasks      *tasks.Store
	settings   *settings.Store
	hosts      *hosts.Store
	knownHosts *sshx.KnownHosts
	creds      *awscred.Resolver
	jobs       *jobs.Pipeline
	options    *optionsCache
}

func New(cfg Config) (*Server, error) {
	if cfg.DataDir == "" {
		return nil, errors.New("data directory is required")
	}
	dir, err := filepath.Abs(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	cfg.DataDir = dir
	password, err := datadir.Init(dir)
	if err != nil {
		return nil, err
	}
	v, err := vault.Open(dir, password)
	if err != nil {
		return nil, err
	}
	t, err := tasks.Open(filepath.Join(dir, datadir.TasksFile))
	if err != nil {
		return nil, err
	}
	st, err := settings.Open(filepath.Join(dir, datadir.SettingsFile))
	if err != nil {
		return nil, err
	}
	h, err := hosts.Open(filepath.Join(dir, datadir.HostsFile))
	if err != nil {
		return nil, err
	}
	return &Server{
		cfg:        cfg,
		mux:        http.NewServeMux(),
		vault:      v,
		tasks:      t,
		settings:   st,
		hosts:      h,
		knownHosts: sshx.NewKnownHosts(filepath.Join(dir, datadir.KnownHostsFile)),
		creds:      awscred.NewResolver(v, dir),
		jobs:       jobs.New(),
		options:    newOptionsCache(),
	}, nil
}

func (s *Server) Setup() error {
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return err
	}
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	s.mux.HandleFunc("GET /sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		http.ServeFileFS(w, r, staticFS, "sw.js")
	})

	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.routeVault()
	s.routeTasks()
	s.routeSettings()
	s.routeSSH()
	s.mux.HandleFunc("GET /ws/terminal", s.handleTerminal)
	s.routeAWS()
	s.routeMachines()
	s.routeTools()

	s.mux.HandleFunc("/api/", s.handleAPINotFound)
	s.mux.HandleFunc("/ws/", http.NotFound)
	s.mux.HandleFunc("/", s.handleIndex)
	return nil
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	log.Info().Str("addr", addr).Str("data", s.cfg.DataDir).Msg("starting")
	return http.ListenAndServe(addr, http.NewCrossOriginProtection().Handler(s.mux))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAPINotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "no route for "+r.Method+" "+r.URL.Path)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(data)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.MarshalWrite(w, v); err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) error {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		return fmt.Errorf("%w: %v", errBadRequest, err)
	}
	if len(data) == 0 {
		return fmt.Errorf("%w: request body is empty", errBadRequest)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("%w: invalid request body: %v", errBadRequest, err)
	}
	return nil
}

func badRequest(format string, args ...any) error {
	return fmt.Errorf("%w: %s", errBadRequest, fmt.Sprintf(format, args...))
}

func statusOf(err error, fallback int) int {
	if _, ok := errors.AsType[*vault.ReferencedError](err); ok {
		return http.StatusConflict
	}
	if _, ok := errors.AsType[*machine.NotRunningError](err); ok {
		return http.StatusConflict
	}
	is := func(targets ...error) bool {
		for _, t := range targets {
			if errors.Is(err, t) {
				return true
			}
		}
		return false
	}
	switch {
	case is(awscred.ErrLoginRequired):
		return http.StatusUnauthorized
	case is(errBadRequest, vault.ErrInvalid, tasks.ErrInvalid, hosts.ErrInvalid, settings.ErrInvalid,
		awscred.ErrInvalidRef, awsrun.ErrInvalid):
		return http.StatusBadRequest
	case is(vault.ErrNotFound, tasks.ErrNotFound, hosts.ErrNotFound, jobs.ErrNotFound, awscred.ErrNotFound, machine.ErrNotFound):
		return http.StatusNotFound
	case is(vault.ErrConflict, hosts.ErrConflict, scaffold.ErrKeyMismatch, machine.ErrNoKey):
		return http.StatusConflict
	case is(awsrun.ErrNotInstalled):
		return http.StatusNotImplemented
	case is(machine.ErrSSH):
		return http.StatusBadGateway
	}
	return fallback
}

func errMessage(err error) string {
	if errors.Is(err, awscred.ErrLoginRequired) {
		return awscred.ErrLoginRequired.Error()
	}
	return err.Error()
}

func (s *Server) fail(w http.ResponseWriter, r *http.Request, err error) {
	s.failWith(w, r, err, http.StatusInternalServerError)
}

func (s *Server) failRemote(w http.ResponseWriter, r *http.Request, err error) {
	s.failWith(w, r, err, http.StatusBadGateway)
}

func (s *Server) failWith(w http.ResponseWriter, r *http.Request, err error, fallback int) {
	status := statusOf(err, fallback)
	msg := errMessage(err)
	event := log.Debug()
	if status >= 500 {
		event = log.Error()
	}
	event.Err(err).Str("method", r.Method).Str("path", r.URL.Path).Int("status", status).Msg("request failed")
	if ref, ok := errors.AsType[*vault.ReferencedError](err); ok {
		writeJSON(w, status, map[string]any{"error": "secret is referenced", "used_by": ref.UsedBy})
		return
	}
	writeError(w, status, msg)
}
