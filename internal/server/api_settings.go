package server

import (
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/tanq16/anbu/internal/settings"
)

func (s *Server) routeSettings() {
	s.mux.HandleFunc("GET /api/settings", s.handleGetSettings)
	s.mux.HandleFunc("PUT /api/settings", s.handlePutSettings)
	s.mux.HandleFunc("PUT /api/settings/password", s.handleRotatePassword)
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.settings.Get())
}

func (s *Server) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		s.fail(w, r, badRequest("%v", err))
		return
	}
	next, err := settings.Decode(data)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.settings.Put(next); err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, next)
}

func (s *Server) handleRotatePassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Password string `json:"password"`
	}
	if err := readJSON(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.vault.Rotate(in.Password); err != nil {
		s.fail(w, r, err)
		return
	}
	log.Info().Msg("vault password rotated")
	w.WriteHeader(http.StatusNoContent)
}
