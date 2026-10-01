package vault

import (
	"cmp"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"uuid"

	"github.com/tanq16/anbu/internal/datadir"
)

type Store struct {
	mu          sync.RWMutex
	root        string
	key         []byte
	salt        []byte
	iter        int
	secrets     map[string]Secret
	machineKeys map[string]MachineKey
}

func Open(root, password string) (*Store, error) {
	if err := recoverRotation(root, password); err != nil {
		return nil, err
	}
	s := &Store{root: root}
	data, err := os.ReadFile(s.path(datadir.VaultFile))
	if errors.Is(err, fs.ErrNotExist) {
		s.salt = newSalt()
		s.iter = kdfIterations
		if s.key, err = deriveKey(password, s.salt, s.iter); err != nil {
			return nil, err
		}
		s.secrets, s.machineKeys = map[string]Secret{}, map[string]MachineKey{}
		return s, s.persist()
	}
	if err != nil {
		return nil, err
	}
	o, err := openVault(data, password)
	if err != nil {
		return nil, err
	}
	s.secrets, s.machineKeys, s.key, s.salt, s.iter = o.Secrets, o.MachineKeys, o.key, o.salt, o.iter
	return s, nil
}

func recoverRotation(root, password string) error {
	passwordNext := filepath.Join(root, datadir.PasswordFile+".next")
	if err := os.Remove(passwordNext); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	vaultNext := filepath.Join(root, datadir.VaultFile+".next")
	data, err := os.ReadFile(vaultNext)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := openVault(data, password); err == nil {
		return os.Rename(vaultNext, filepath.Join(root, datadir.VaultFile))
	}
	return os.Remove(vaultNext)
}

func (s *Store) path(name string) string {
	return filepath.Join(s.root, name)
}

func (s *Store) persist() error {
	data, err := sealVault(s.contents(), s.key, s.salt, s.iter)
	if err != nil {
		return err
	}
	return datadir.WriteFile(s.path(datadir.VaultFile), data)
}

func (s *Store) contents() contents {
	return contents{Secrets: s.secrets, MachineKeys: s.machineKeys}
}

func (s *Store) sorted() []Secret {
	out := slices.Collect(maps.Values(s.secrets))
	slices.SortFunc(out, func(a, b Secret) int {
		return cmp.Or(cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), cmp.Compare(a.ID, b.ID))
	})
	return out
}

func (s *Store) List() []Secret {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sorted()
}

func (s *Store) find(ref string) (Secret, bool) {
	if sec, ok := s.secrets[ref]; ok {
		return sec, true
	}
	for _, sec := range s.secrets {
		if sec.Name == ref {
			return sec, true
		}
	}
	return Secret{}, false
}

func (s *Store) Get(ref string) (Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sec, ok := s.find(ref)
	if !ok {
		return Secret{}, ErrNotFound
	}
	return sec, nil
}

func (s *Store) nameTaken(name, exceptID string) bool {
	for _, sec := range s.secrets {
		if sec.Name == name && sec.ID != exceptID {
			return true
		}
	}
	return false
}

func (s *Store) Create(in Secret) (Secret, error) {
	if err := in.normalize(); err != nil {
		return Secret{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.insert(in)
}

func (s *Store) insert(in Secret) (Secret, error) {
	if s.nameTaken(in.Name, "") {
		return Secret{}, conflict("a secret named %s already exists", in.Name)
	}
	now := time.Now().UTC()
	in.ID = uuid.NewV7().String()
	in.CreatedAt, in.UpdatedAt = now, now
	s.secrets[in.ID] = in
	if err := s.persist(); err != nil {
		delete(s.secrets, in.ID)
		return Secret{}, err
	}
	return in, nil
}

func (s *Store) GenerateSSHKey(name string) (Secret, error) {
	fields, err := newSSHKeyFields(strings.TrimSpace(name))
	if err != nil {
		return Secret{}, err
	}
	return s.Create(Secret{Name: name, Type: TypeSSHKey, Fields: fields})
}

func (s *Store) Update(ref string, in Secret) (Secret, error) {
	if err := in.normalize(); err != nil {
		return Secret{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.find(ref)
	if !ok {
		return Secret{}, ErrNotFound
	}
	if in.Type != cur.Type {
		return Secret{}, invalid("type cannot change from %s to %s", cur.Type, in.Type)
	}
	if s.nameTaken(in.Name, cur.ID) {
		return Secret{}, conflict("a secret named %s already exists", in.Name)
	}
	in.ID, in.CreatedAt, in.UpdatedAt = cur.ID, cur.CreatedAt, time.Now().UTC()
	s.secrets[in.ID] = in
	if err := s.persist(); err != nil {
		s.secrets[cur.ID] = cur
		return Secret{}, err
	}
	return in, nil
}

func (s *Store) Delete(ref string, usedBy func(Secret) []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sec, ok := s.find(ref)
	if !ok {
		return ErrNotFound
	}
	var refs []string
	if usedBy != nil {
		refs = usedBy(sec)
	}
	if len(refs) > 0 {
		return &ReferencedError{UsedBy: refs}
	}
	return s.remove(sec)
}

func (s *Store) remove(sec Secret) error {
	delete(s.secrets, sec.ID)
	if err := s.persist(); err != nil {
		s.secrets[sec.ID] = sec
		return err
	}
	return nil
}

func (s *Store) Rotate(password string) error {
	if password == "" {
		return invalid("password is required")
	}
	if strings.ContainsAny(password, "\r\n") {
		return invalid("password cannot contain a line break")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	salt := newSalt()
	key, err := deriveKey(password, salt, kdfIterations)
	if err != nil {
		return err
	}
	data, err := sealVault(s.contents(), key, salt, kdfIterations)
	if err != nil {
		return err
	}
	vaultNext := s.path(datadir.VaultFile + ".next")
	passwordNext := s.path(datadir.PasswordFile + ".next")
	if err := datadir.WriteFile(vaultNext, data); err != nil {
		return err
	}
	if err := datadir.WriteFile(passwordNext, []byte(password+"\n")); err != nil {
		os.Remove(vaultNext)
		return err
	}
	if err := os.Rename(passwordNext, s.path(datadir.PasswordFile)); err != nil {
		os.Remove(vaultNext)
		os.Remove(passwordNext)
		return err
	}
	s.key, s.salt, s.iter = key, salt, kdfIterations
	if err := os.Rename(vaultNext, s.path(datadir.VaultFile)); err != nil {
		if err := s.persist(); err != nil {
			return err
		}
		os.Remove(vaultNext)
	}
	return nil
}
