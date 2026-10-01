package server

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tanq16/anbu/internal/awscred"
	"github.com/tanq16/anbu/internal/awsx"
	"github.com/tanq16/anbu/internal/hosts"
	"github.com/tanq16/anbu/internal/machine"
	"github.com/tanq16/anbu/internal/sshx"
	"github.com/tanq16/anbu/internal/vault"
	"golang.org/x/crypto/ssh"
)

const profileTimeout = 30 * time.Second

type sshTarget struct {
	Kind      string         `json:"kind"`
	Ref       string         `json:"ref"`
	Name      string         `json:"name"`
	ID        string         `json:"id,omitempty"`
	Addr      string         `json:"addr"`
	Address   string         `json:"address,omitempty"`
	Port      int            `json:"port,omitzero"`
	PublicIP  string         `json:"public_ip,omitempty"`
	Profile   string         `json:"profile,omitempty"`
	Account   string         `json:"account,omitempty"`
	Region    string         `json:"region,omitempty"`
	User      string         `json:"user"`
	KeySecret string         `json:"key_secret,omitempty"`
	Key       string         `json:"key"`
	Alias     string         `json:"alias"`
	State     string         `json:"state"`
	Target    sshx.TargetRef `json:"target"`
}

type profileError struct {
	Profile string `json:"profile"`
	Status  int    `json:"status"`
	Error   string `json:"error"`
}

func (s *Server) routeSSH() {
	s.mux.HandleFunc("GET /api/hosts", s.handleListHosts)
	s.mux.HandleFunc("POST /api/hosts", s.handleCreateHost)
	s.mux.HandleFunc("PUT /api/hosts/{ref}", s.handleUpdateHost)
	s.mux.HandleFunc("DELETE /api/hosts/{ref}", s.handleDeleteHost)
	s.mux.HandleFunc("GET /api/ssh/targets", s.handleSSHTargets)
	s.mux.HandleFunc("DELETE /api/ssh/known-hosts/{alias}", s.handleForgetHostKey)
}

func (s *Server) handleListHosts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.hosts.List())
}

func (s *Server) hostInput(w http.ResponseWriter, r *http.Request) (hosts.Host, error) {
	var in hosts.Host
	if err := readJSON(w, r, &in); err != nil {
		return hosts.Host{}, err
	}
	if in.KeySecret == "" {
		return in, nil
	}
	sec, err := s.vault.Get(in.KeySecret)
	if errors.Is(err, vault.ErrNotFound) {
		return hosts.Host{}, badRequest("key_secret %q is not a secret in the vault", in.KeySecret)
	}
	if err != nil {
		return hosts.Host{}, err
	}
	if sec.Type != vault.TypeSSHKey {
		return hosts.Host{}, badRequest("key_secret %s has type %s, not %s", sec.Name, sec.Type, vault.TypeSSHKey)
	}
	in.KeySecret = sec.ID
	return in, nil
}

func (s *Server) handleCreateHost(w http.ResponseWriter, r *http.Request) {
	in, err := s.hostInput(w, r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	h, err := s.hosts.Create(in)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	log.Info().Str("host", h.Name).Msg("host created")
	writeJSON(w, http.StatusCreated, h)
}

func (s *Server) handleUpdateHost(w http.ResponseWriter, r *http.Request) {
	in, err := s.hostInput(w, r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	h, err := s.hosts.Update(r.PathValue("ref"), in)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	log.Info().Str("host", h.Name).Msg("host updated")
	writeJSON(w, http.StatusOK, h)
}

func (s *Server) handleDeleteHost(w http.ResponseWriter, r *http.Request) {
	if err := s.hosts.Delete(r.PathValue("ref")); err != nil {
		s.fail(w, r, err)
		return
	}
	log.Info().Str("host", r.PathValue("ref")).Msg("host deleted")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleForgetHostKey(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("alias")
	if alias == "" {
		s.fail(w, r, badRequest("alias is required"))
		return
	}
	removed, err := s.knownHosts.Remove(alias)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	log.Info().Str("alias", alias).Int("removed", removed).Msg("host key forgotten")
	writeJSON(w, http.StatusOK, map[string]int{"removed": removed})
}

func (s *Server) handleSSHTargets(w http.ResponseWriter, r *http.Request) {
	targets := []sshTarget{}
	for _, h := range s.hosts.List() {
		var keyName string
		if sec, err := s.vault.Get(h.KeySecret); err == nil {
			keyName = sec.Name
		}
		targets = append(targets, sshTarget{
			Kind:      "host",
			Ref:       h.Name,
			Name:      h.Name,
			ID:        h.ID,
			Addr:      h.Alias(),
			Address:   h.Address,
			Port:      h.Port,
			User:      h.User,
			KeySecret: h.KeySecret,
			Key:       keyName,
			Alias:     h.Alias(),
			State:     "ready",
			Target:    sshx.TargetRef{Host: h.ID},
		})
	}

	profiles := s.creds.Profiles()
	type outcome struct {
		clients   *awsx.Clients
		instances []awsx.Instance
		err       error
	}
	results := make([]outcome, len(profiles))
	var wg sync.WaitGroup
	for i, p := range profiles {
		wg.Go(func() {
			ctx, cancel := context.WithTimeout(r.Context(), profileTimeout)
			defer cancel()
			c, err := s.clients(ctx, awscred.Source{Profile: p.Ref})
			if err != nil {
				results[i].err = err
				return
			}
			instances, err := c.ManagedInstances(ctx)
			results[i] = outcome{clients: c, instances: instances, err: err}
		})
	}
	wg.Wait()

	errs := []profileError{}
	seen := map[string]bool{}
	for i, res := range results {
		if res.err != nil {
			status := statusOf(res.err, http.StatusBadGateway)
			log.Debug().Err(res.err).Str("profile", profiles[i].Ref).Msg("profile failed while listing ssh targets")
			errs = append(errs, profileError{Profile: profiles[i].Ref, Status: status, Error: errMessage(res.err)})
			continue
		}
		c := res.clients
		place := c.Account + "/" + c.Region
		if seen[place] {
			continue
		}
		seen[place] = true
		for _, inst := range res.instances {
			targets = append(targets, sshTarget{
				Kind:     "machine",
				Ref:      c.Profile + "/" + inst.Name,
				Name:     inst.Name,
				Addr:     inst.PublicIP,
				PublicIP: inst.PublicIP,
				Profile:  c.Profile,
				Account:  c.Account,
				Region:   c.Region,
				User:     machine.SSHUser,
				Key:      "scaffold " + c.Account + "/" + c.Region,
				Alias:    c.HostKeyAlias(inst.Name),
				State:    inst.State,
				Target:   sshx.TargetRef{Machine: &sshx.MachineRef{Profile: c.Profile, Name: inst.Name}},
			})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"targets": targets, "errors": errs})
}

func (s *Server) dialTarget(ctx context.Context, t sshx.TargetRef) (*ssh.Client, error) {
	ep, err := s.endpoint(ctx, t)
	if err != nil {
		return nil, err
	}
	return sshx.Dial(ctx, ep, s.knownHosts)
}

func (s *Server) endpoint(ctx context.Context, t sshx.TargetRef) (sshx.Endpoint, error) {
	if t.Machine != nil {
		c, err := s.clients(ctx, awscred.Source{Profile: t.Machine.Profile})
		if err != nil {
			return sshx.Endpoint{}, err
		}
		return machine.Endpoint(ctx, c, s.vault, t.Machine.Name)
	}
	if t.Host == "" {
		return sshx.Endpoint{}, badRequest("target names neither a host nor a machine")
	}
	h, err := s.hosts.Get(t.Host)
	if err != nil {
		return sshx.Endpoint{}, err
	}
	sec, err := s.vault.Get(h.KeySecret)
	if err != nil {
		return sshx.Endpoint{}, err
	}
	signer, err := vault.Signer(sec)
	if err != nil {
		return sshx.Endpoint{}, err
	}
	return sshx.Endpoint{Addr: h.Addr(), User: h.User, Signer: signer, Alias: h.Alias()}, nil
}
