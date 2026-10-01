package vault

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"
)

var (
	ErrNotFound = errors.New("secret not found")
	ErrInvalid  = errors.New("invalid secret")
	ErrConflict = errors.New("secret conflict")
)

type ReferencedError struct {
	UsedBy []string
}

func (e *ReferencedError) Error() string {
	return "secret is referenced by " + strings.Join(e.UsedBy, ", ")
}

type Secret struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      Type              `json:"type"`
	Fields    map[string]string `json:"fields"`
	Profiles  []AWSProfile      `json:"profiles,omitempty"`
	TOTP      *TOTP             `json:"totp,omitempty"`
	Custom    []CustomField     `json:"custom,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type CustomField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Hidden bool   `json:"hidden"`
}

type AWSProfile struct {
	Name      string        `json:"name"`
	AccountID string        `json:"account_id"`
	RoleName  string        `json:"role_name"`
	Region    string        `json:"region"`
	Custom    []CustomField `json:"custom,omitempty"`
}

type Type string

const (
	TypeLogin     Type = "login"
	TypeSSHKey    Type = "ssh-key"
	TypeAWSStatic Type = "aws-static"
	TypeAWSSSO    Type = "aws-sso"
	TypeGitHubPAT Type = "github-pat"
	TypeGeneric   Type = "generic"
)

type Kind string

const (
	KindText      Kind = "text"
	KindSecret    Kind = "secret"
	KindURL       Kind = "url"
	KindMultiline Kind = "multiline-secret"
	KindDate      Kind = "date"
)

type FieldSpec struct {
	Key      string `json:"key"`
	Kind     Kind   `json:"kind"`
	Required bool   `json:"required"`
}

type TypeSpec struct {
	Label    string      `json:"label"`
	Fields   []FieldSpec `json:"fields"`
	Profiles bool        `json:"profiles,omitzero"`
}

var Schema = map[Type]TypeSpec{
	TypeLogin: {Label: "Login", Fields: []FieldSpec{
		{"username", KindText, false},
		{"password", KindSecret, false},
		{"url", KindURL, false},
	}},
	TypeSSHKey: {Label: "SSH key", Fields: []FieldSpec{
		{"private_key", KindMultiline, true},
		{"public_key", KindText, false},
		{"passphrase", KindSecret, false},
	}},
	TypeAWSStatic: {Label: "AWS static", Fields: []FieldSpec{
		{"access_key_id", KindText, true},
		{"secret_access_key", KindSecret, true},
		{"session_token", KindSecret, false},
		{"region", KindText, true},
	}},
	TypeAWSSSO: {Label: "AWS SSO", Profiles: true, Fields: []FieldSpec{
		{"start_url", KindURL, true},
		{"sso_region", KindText, true},
	}},
	TypeGitHubPAT: {Label: "GitHub PAT", Fields: []FieldSpec{
		{"token", KindSecret, true},
		{"username", KindText, false},
		{"expires_at", KindDate, false},
	}},
	TypeGeneric: {Label: "Generic", Fields: []FieldSpec{
		{"value", KindMultiline, false},
	}},
}

const ScaffoldKeyPrefix = "sharingan-"

func (s Secret) VisibleFields() map[string]string {
	out := map[string]string{}
	for _, f := range Schema[s.Type].Fields {
		if f.Kind == KindText || f.Kind == KindURL || f.Kind == KindDate {
			out[f.Key] = s.Fields[f.Key]
		}
	}
	return out
}

func (s Secret) Profile(name string) (AWSProfile, bool) {
	i := slices.IndexFunc(s.Profiles, func(p AWSProfile) bool { return p.Name == name })
	if i < 0 {
		return AWSProfile{}, false
	}
	return s.Profiles[i], true
}

func validName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if name != strings.TrimSpace(name) {
		return errors.New("name has leading or trailing spaces")
	}
	if strings.ContainsAny(name, "/:") {
		return errors.New(`name cannot contain "/" or ":"`)
	}
	if strings.ContainsFunc(name, func(r rune) bool { return r < 0x20 || r == 0x7f }) {
		return errors.New("name cannot contain control characters")
	}
	return nil
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

func conflict(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrConflict, fmt.Sprintf(format, args...))
}

func (s *Secret) normalize() error {
	s.Name = strings.TrimSpace(s.Name)
	if err := validName(s.Name); err != nil {
		return invalid("%v", err)
	}
	spec, ok := Schema[s.Type]
	if !ok {
		return invalid("unknown type %q", s.Type)
	}

	fields := make(map[string]string, len(spec.Fields))
	for _, f := range spec.Fields {
		fields[f.Key] = s.Fields[f.Key]
	}
	for key := range maps.Keys(s.Fields) {
		if _, known := fields[key]; !known {
			return invalid("unknown field %q for type %s", key, s.Type)
		}
	}
	for _, f := range spec.Fields {
		value := fields[f.Key]
		if f.Kind != KindSecret && f.Kind != KindMultiline {
			value = strings.TrimSpace(value)
			fields[f.Key] = value
		}
		if f.Required && strings.TrimSpace(value) == "" {
			return invalid("field %q is required for type %s", f.Key, s.Type)
		}
		if f.Kind == KindDate && value != "" {
			if _, err := time.Parse(time.DateOnly, value); err != nil {
				return invalid("field %q must be a YYYY-MM-DD date", f.Key)
			}
		}
	}
	s.Fields = fields

	if err := s.normalizeProfiles(spec); err != nil {
		return err
	}
	if err := normalizeCustom(s.Custom); err != nil {
		return err
	}
	if err := s.normalizeTOTP(); err != nil {
		return err
	}
	if s.Type == TypeSSHKey {
		return s.fillPublicKey()
	}
	return nil
}

func (s *Secret) normalizeProfiles(spec TypeSpec) error {
	if !spec.Profiles {
		if len(s.Profiles) > 0 {
			return invalid("profiles are only allowed on %s secrets", TypeAWSSSO)
		}
		s.Profiles = nil
		return nil
	}
	if len(s.Profiles) == 0 {
		return invalid("an %s secret needs at least one profile", TypeAWSSSO)
	}
	seen := map[string]bool{}
	for i := range s.Profiles {
		p := &s.Profiles[i]
		p.Name = strings.TrimSpace(p.Name)
		p.AccountID = strings.TrimSpace(p.AccountID)
		p.RoleName = strings.TrimSpace(p.RoleName)
		p.Region = strings.TrimSpace(p.Region)
		if err := validName(p.Name); err != nil {
			return invalid("profile %v", err)
		}
		if seen[p.Name] {
			return invalid("duplicate profile %q", p.Name)
		}
		seen[p.Name] = true
		if p.AccountID == "" || p.RoleName == "" || p.Region == "" {
			return invalid("profile %q needs account_id, role_name, and region", p.Name)
		}
		if err := normalizeCustom(p.Custom); err != nil {
			return err
		}
	}
	return nil
}

func normalizeCustom(custom []CustomField) error {
	for i := range custom {
		custom[i].Name = strings.TrimSpace(custom[i].Name)
		if custom[i].Name == "" {
			return invalid("custom field name is required")
		}
	}
	return nil
}

func (s *Secret) normalizeTOTP() error {
	if s.TOTP == nil || s.TOTP.Secret == "" {
		s.TOTP = nil
		return nil
	}
	t := *s.TOTP
	if t.Algorithm == "" {
		t.Algorithm = "SHA1"
	}
	if t.Digits == 0 {
		t.Digits = 6
	}
	if t.Period == 0 {
		t.Period = 30
	}
	if err := t.normalize(); err != nil {
		return invalid("totp %v", err)
	}
	s.TOTP = &t
	return nil
}
