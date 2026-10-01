package server

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/credentials/processcreds"
	"github.com/rs/zerolog/log"
	"github.com/tanq16/anbu/internal/awscred"
	"github.com/tanq16/anbu/internal/vault"
)

const maxImportBody = 32 << 20

type secretView struct {
	vault.Secret
	UsedBy []string `json:"used_by"`
}

type secretListItem struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      vault.Type        `json:"type"`
	HasTOTP   bool              `json:"has_totp"`
	UpdatedAt time.Time         `json:"updated_at"`
	Fields    map[string]string `json:"fields"`
	Profiles  []profileListItem `json:"profiles,omitzero"`
}

type profileListItem struct {
	Name      string `json:"name"`
	Ref       string `json:"ref"`
	AccountID string `json:"account_id"`
	RoleName  string `json:"role_name"`
	Region    string `json:"region"`
}

type totpView struct {
	Code      string `json:"code"`
	Period    int    `json:"period"`
	Remaining int    `json:"remaining"`
}

func (s *Server) routeVault() {
	s.mux.HandleFunc("GET /api/secret-types", s.handleSecretTypes)
	s.mux.HandleFunc("GET /api/secrets", s.handleListSecrets)
	s.mux.HandleFunc("POST /api/secrets", s.handleCreateSecret)
	s.mux.HandleFunc("POST /api/secrets/generate-ssh-key", s.handleGenerateSSHKey)
	s.mux.HandleFunc("GET /api/secrets/{ref}", s.handleGetSecret)
	s.mux.HandleFunc("PUT /api/secrets/{ref}", s.handleUpdateSecret)
	s.mux.HandleFunc("DELETE /api/secrets/{ref}", s.handleDeleteSecret)
	s.mux.HandleFunc("GET /api/secrets/{ref}/totp", s.handleTOTP)
	s.mux.HandleFunc("GET /api/vault/export", s.handleExport)
	s.mux.HandleFunc("POST /api/vault/import", s.handleImport)
}

func (s *Server) usedBy(sec vault.Secret) []string {
	refs := s.hosts.Referencing(sec.ID)
	if refs == nil {
		refs = []string{}
	}
	return refs
}

func (s *Server) handleSecretTypes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, vault.Schema)
}

func (s *Server) handleListSecrets(w http.ResponseWriter, r *http.Request) {
	list := s.vault.List()
	out := make([]secretListItem, 0, len(list))
	for _, sec := range list {
		item := secretListItem{
			ID:        sec.ID,
			Name:      sec.Name,
			Type:      sec.Type,
			HasTOTP:   sec.TOTP != nil,
			UpdatedAt: sec.UpdatedAt,
			Fields:    sec.VisibleFields(),
		}
		for _, p := range sec.Profiles {
			item.Profiles = append(item.Profiles, profileListItem{
				Name:      p.Name,
				Ref:       sec.Name + ":" + p.Name,
				AccountID: p.AccountID,
				RoleName:  p.RoleName,
				Region:    p.Region,
			})
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetSecret(w http.ResponseWriter, r *http.Request) {
	ref := r.PathValue("ref")
	if strings.Contains(ref, ":") {
		s.writeProcessCreds(w, r, ref)
		return
	}
	sec, err := s.vault.Get(ref)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, secretView{Secret: sec, UsedBy: s.usedBy(sec)})
}

func (s *Server) writeProcessCreds(w http.ResponseWriter, r *http.Request, ref string) {
	creds, _, err := s.creds.Credentials(r.Context(), awscred.Source{Profile: ref})
	if err != nil {
		s.failRemote(w, r, err)
		return
	}
	resp := processcreds.CredentialProcessResponse{
		Version:         1,
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		AccountID:       creds.AccountID,
	}
	if creds.CanExpire {
		expires := creds.Expires.UTC()
		resp.Expiration = &expires
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleCreateSecret(w http.ResponseWriter, r *http.Request) {
	var in vault.Secret
	if err := readJSON(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	sec, err := s.vault.Create(in)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	log.Info().Str("secret", sec.Name).Str("type", string(sec.Type)).Msg("secret created")
	writeJSON(w, http.StatusCreated, sec)
}

func (s *Server) handleGenerateSSHKey(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	if err := readJSON(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	sec, err := s.vault.GenerateSSHKey(in.Name)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	log.Info().Str("secret", sec.Name).Msg("ssh key generated")
	writeJSON(w, http.StatusCreated, sec)
}

func (s *Server) handleUpdateSecret(w http.ResponseWriter, r *http.Request) {
	var in vault.Secret
	if err := readJSON(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	sec, err := s.vault.Update(r.PathValue("ref"), in)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	log.Info().Str("secret", sec.Name).Msg("secret updated")
	writeJSON(w, http.StatusOK, sec)
}

func (s *Server) handleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	if err := s.vault.Delete(r.PathValue("ref"), s.usedBy); err != nil {
		s.fail(w, r, err)
		return
	}
	log.Info().Str("secret", r.PathValue("ref")).Msg("secret deleted")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleTOTP(w http.ResponseWriter, r *http.Request) {
	sec, err := s.vault.Get(r.PathValue("ref"))
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if sec.TOTP == nil {
		writeError(w, http.StatusNotFound, "secret "+sec.Name+" has no totp")
		return
	}
	now := time.Now()
	code, err := sec.TOTP.Code(now)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	period := sec.TOTP.Period
	writeJSON(w, http.StatusOK, totpView{Code: code, Period: period, Remaining: period - int(now.Unix()%int64(period))})
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	doc := s.vault.Export()
	data, err := json.Marshal(doc, jsontext.WithIndent("  "))
	if err != nil {
		s.fail(w, r, err)
		return
	}
	name := "anbu-vault-" + doc.ExportedAt.Format("20060102") + ".json"
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.Write(append(data, '\n'))
	log.Info().Int("secrets", len(doc.Secrets)).Msg("vault exported")
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxImportBody)
	var doc vault.Export
	if err := json.UnmarshalRead(r.Body, &doc); err != nil {
		s.fail(w, r, badRequest("invalid import document: %v", err))
		return
	}
	added, skipped, err := s.vault.Import(doc)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	log.Info().Int("added", len(added)).Int("skipped", len(skipped)).Msg("vault imported")
	writeJSON(w, http.StatusOK, map[string][]string{"added": added, "skipped": skipped})
}
