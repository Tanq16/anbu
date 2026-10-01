package sshx

import (
	"errors"
	"strings"

	"golang.org/x/crypto/ssh"
)

type TargetRef struct {
	Host    string      `json:"host,omitempty"`
	Machine *MachineRef `json:"machine,omitempty"`
}

type MachineRef struct {
	Profile string `json:"profile"`
	Name    string `json:"name"`
}

type Endpoint struct {
	Addr   string
	User   string
	Signer ssh.Signer
	Alias  string
}

func ParseTarget(s string) (TargetRef, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return TargetRef{}, errors.New("target is required")
	}
	profile, name, found := strings.Cut(s, "/")
	if !found {
		return TargetRef{Host: s}, nil
	}
	if profile == "" || name == "" {
		return TargetRef{}, errors.New("a machine target is <profile-ref>/<machine>")
	}
	return TargetRef{Machine: &MachineRef{Profile: profile, Name: name}}, nil
}

func (t TargetRef) String() string {
	if t.Machine != nil {
		return t.Machine.Profile + "/" + t.Machine.Name
	}
	return t.Host
}
