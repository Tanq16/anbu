package server

import (
	"cmp"
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tanq16/anbu/internal/awscred"
	"github.com/tanq16/anbu/internal/awsrun"
	"github.com/tanq16/anbu/internal/awsx"
	"github.com/tanq16/anbu/internal/vault"
)

func (s *Server) routeAWS() {
	s.mux.HandleFunc("POST /api/aws/sso/{secret}/login", s.handleSSOLogin)
	s.mux.HandleFunc("GET /api/aws/sso/{secret}/status", s.handleSSOStatus)
	s.mux.HandleFunc("GET /api/aws/sso/{secret}/accounts", s.handleSSOAccounts)
	s.mux.HandleFunc("GET /api/aws/whoami", s.handleWhoami)
	s.mux.HandleFunc("POST /api/aws/run", s.handleAWSRun)
	s.mux.HandleFunc("GET /api/aws/jobs", s.handleListJobs)
	s.mux.HandleFunc("GET /api/aws/jobs/{id}", s.handleGetJob)
}

func (s *Server) clients(ctx context.Context, src awscred.Source) (*awsx.Clients, error) {
	cfg, err := s.creds.Config(ctx, src)
	if err != nil {
		return nil, err
	}
	c, err := awsx.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	c.Profile = src.Label()
	return c, nil
}

func querySource(r *http.Request) awscred.Source {
	return awscred.Source{Profile: r.URL.Query().Get("profile")}
}

func (s *Server) ssoSecret(r *http.Request) (vault.Secret, error) {
	sec, err := s.vault.Get(r.PathValue("secret"))
	if err != nil {
		return vault.Secret{}, err
	}
	if sec.Type != vault.TypeAWSSSO {
		return vault.Secret{}, badRequest("%s has type %s, not %s", sec.Name, sec.Type, vault.TypeAWSSSO)
	}
	return sec, nil
}

func (s *Server) handleSSOLogin(w http.ResponseWriter, r *http.Request) {
	sec, err := s.ssoSecret(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	prompt, err := s.creds.StartLogin(r.Context(), sec)
	if err != nil {
		s.failRemote(w, r, err)
		return
	}
	log.Info().Str("secret", sec.Name).Msg("sso login started")
	writeJSON(w, http.StatusOK, prompt)
}

func (s *Server) handleSSOStatus(w http.ResponseWriter, r *http.Request) {
	sec, err := s.ssoSecret(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	status, err := s.creds.Status(sec)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleSSOAccounts(w http.ResponseWriter, r *http.Request) {
	sec, err := s.ssoSecret(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	accounts, err := s.creds.Accounts(r.Context(), sec)
	if err != nil {
		s.failRemote(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, accounts)
}

func (s *Server) handleWhoami(w http.ResponseWriter, r *http.Request) {
	src := querySource(r)
	if err := s.creds.Validate(src); err != nil {
		s.fail(w, r, err)
		return
	}
	c, err := s.clients(r.Context(), src)
	if err != nil {
		s.failRemote(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"account": c.Account,
		"arn":     c.ARN,
		"region":  c.Region,
		"profile": c.Profile,
	})
}

func (s *Server) handleAWSRun(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Source  awscred.Source `json:"source"`
		Command string         `json:"command"`
		Region  string         `json:"region"`
	}
	if err := readJSON(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	if in.Source.Profile == "" && in.Source.Inline == nil {
		in.Source = querySource(r)
	}
	if err := s.creds.Validate(in.Source); err != nil {
		s.fail(w, r, err)
		return
	}
	if strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(in.Command), "aws")) == "" {
		s.fail(w, r, badRequest("command is required"))
		return
	}
	if _, err := awsrun.Split(in.Command); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := awsrun.Installed(); err != nil {
		s.fail(w, r, err)
		return
	}
	creds, region, err := s.creds.Credentials(r.Context(), in.Source)
	if err != nil {
		s.failRemote(w, r, err)
		return
	}
	timeout := time.Duration(s.settings.Get().AWSCommandTimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()
	res, err := awsrun.Run(ctx, s.cfg.DataDir, creds, cmp.Or(in.Region, region), in.Command)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		res.Stderr += "\nanbu: command timed out after " + timeout.String()
	}
	log.Info().Str("profile", in.Source.Label()).Int("exit_code", res.ExitCode).Msg("aws command ran")
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.jobs.Snapshot())
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	job, err := s.jobs.Get(r.PathValue("id"))
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, job)
}
