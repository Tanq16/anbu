package vault

import (
	"cmp"
	"maps"
	"slices"
	"time"

	"golang.org/x/crypto/ssh"
)

type MachineKey struct {
	Account    string    `json:"account"`
	Region     string    `json:"region"`
	PrivateKey string    `json:"private_key"`
	PublicKey  string    `json:"public_key"`
	CreatedAt  time.Time `json:"created_at"`
}

func machineKeyID(account, region string) string {
	return account + "/" + region
}

func (k MachineKey) Signer() (ssh.Signer, error) {
	return ssh.ParsePrivateKey([]byte(k.PrivateKey))
}

func (k MachineKey) Fingerprint() string {
	pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(k.PublicKey))
	if err != nil {
		return ""
	}
	return ssh.FingerprintSHA256(pub)
}

func (k *MachineKey) normalize() error {
	if k.Account == "" || k.Region == "" {
		return invalid("a machine key needs an account and a region")
	}
	signer, err := k.Signer()
	if err != nil {
		return invalid("private key does not parse: %v", err)
	}
	k.PublicKey = string(ssh.MarshalAuthorizedKey(signer.PublicKey()))
	if k.CreatedAt.IsZero() {
		k.CreatedAt = time.Now().UTC()
	}
	return nil
}

func (s *Store) MachineKey(account, region string) (MachineKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	k, ok := s.machineKeys[machineKeyID(account, region)]
	return k, ok
}

func (s *Store) MachineKeys() []MachineKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.SortedFunc(maps.Values(s.machineKeys), func(a, b MachineKey) int {
		return cmp.Or(cmp.Compare(a.Account, b.Account), cmp.Compare(a.Region, b.Region))
	})
}

func (s *Store) GenerateMachineKey(account, region string) (MachineKey, error) {
	fields, err := newSSHKeyFields("anbu-" + account + "-" + region)
	if err != nil {
		return MachineKey{}, err
	}
	return s.ImportMachineKey(account, region, fields["private_key"])
}

func (s *Store) ImportMachineKey(account, region, privateKey string) (MachineKey, error) {
	k := MachineKey{Account: account, Region: region, PrivateKey: privateKey}
	if err := k.normalize(); err != nil {
		return MachineKey{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := machineKeyID(account, region)
	if _, ok := s.machineKeys[id]; ok {
		return MachineKey{}, conflict("a machine key for %s already exists", id)
	}
	s.machineKeys[id] = k
	if err := s.persist(); err != nil {
		delete(s.machineKeys, id)
		return MachineKey{}, err
	}
	return k, nil
}

func (s *Store) DeleteMachineKey(account, region string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := machineKeyID(account, region)
	k, ok := s.machineKeys[id]
	if !ok {
		return false, nil
	}
	delete(s.machineKeys, id)
	if err := s.persist(); err != nil {
		s.machineKeys[id] = k
		return false, err
	}
	return true, nil
}
