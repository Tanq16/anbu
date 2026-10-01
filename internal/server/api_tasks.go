package server

import (
	"net/http"

	"github.com/tanq16/anbu/internal/tasks"
)

func (s *Server) routeTasks() {
	s.mux.HandleFunc("GET /api/tasks", s.handleListTasks)
	s.mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	s.mux.HandleFunc("PUT /api/tasks/{id}", s.handleUpdateTask)
	s.mux.HandleFunc("DELETE /api/tasks/{id}", s.handleDeleteTask)
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.tasks.List())
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var in tasks.Input
	if err := readJSON(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	in.Done = false
	t, err := s.tasks.Create(in)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var in tasks.Input
	if err := readJSON(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	t, err := s.tasks.Update(r.PathValue("id"), in)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if err := s.tasks.Delete(r.PathValue("id")); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
