package hosts

import (
	"cmp"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"uuid"

	"github.com/tanq16/anbu/internal/datadir"
)

var (
	ErrNotFound = errors.New("host not found")
	ErrInvalid  = errors.New("invalid host")
	ErrConflict = errors.New("host conflict")
)

type Host struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	KeySecret string `json:"key_secret"`
}

func (h Host) Addr() string {
	return net.JoinHostPort(h.Address, strconv.Itoa(h.Port))
}

func (h Host) Alias() string {
	if h.Port == 22 {
		return h.Address
	}
	return "[" + h.Address + "]:" + strconv.Itoa(h.Port)
}

type Store struct {
	mu    sync.Mutex
	path  string
	hosts []Host
}

func Open(path string) (*Store, error) {
	s := &Store{path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.hosts); err != nil {
		return nil, err
	}
	return s, nil
}

func (h *Host) normalize() error {
	h.Name = strings.TrimSpace(h.Name)
	h.Address = strings.TrimSpace(h.Address)
	h.User = strings.TrimSpace(h.User)
	if h.Port == 0 {
		h.Port = 22
	}
	switch {
	case h.Name == "":
		return fmt.Errorf("%w: name is required", ErrInvalid)
	case strings.ContainsAny(h.Name, "/:"):
		return fmt.Errorf(`%w: name cannot contain "/" or ":"`, ErrInvalid)
	case h.Address == "" || strings.ContainsAny(h.Address, " /[]"):
		return fmt.Errorf("%w: address must be a hostname or IP", ErrInvalid)
	case h.Port < 1 || h.Port > 65535:
		return fmt.Errorf("%w: port must be between 1 and 65535", ErrInvalid)
	case h.User == "":
		return fmt.Errorf("%w: user is required", ErrInvalid)
	case h.KeySecret == "":
		return fmt.Errorf("%w: key_secret is required", ErrInvalid)
	}
	return nil
}

func (s *Store) List() []Host {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := slices.Clone(s.hosts)
	slices.SortFunc(out, func(a, b Host) int { return cmp.Compare(a.Name, b.Name) })
	return out
}

func (s *Store) find(ref string) int {
	if i := slices.IndexFunc(s.hosts, func(h Host) bool { return h.ID == ref }); i >= 0 {
		return i
	}
	return slices.IndexFunc(s.hosts, func(h Host) bool { return h.Name == ref })
}

func (s *Store) Get(ref string) (Host, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.find(ref)
	if i < 0 {
		return Host{}, ErrNotFound
	}
	return s.hosts[i], nil
}

func (s *Store) nameTaken(name, exceptID string) bool {
	return slices.ContainsFunc(s.hosts, func(h Host) bool { return h.Name == name && h.ID != exceptID })
}

func (s *Store) Create(h Host) (Host, error) {
	if err := h.normalize(); err != nil {
		return Host{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.nameTaken(h.Name, "") {
		return Host{}, fmt.Errorf("%w: a host named %s already exists", ErrConflict, h.Name)
	}
	h.ID = uuid.NewV7().String()
	s.hosts = append(s.hosts, h)
	if err := s.persist(); err != nil {
		s.hosts = s.hosts[:len(s.hosts)-1]
		return Host{}, err
	}
	return h, nil
}

func (s *Store) Update(ref string, h Host) (Host, error) {
	if err := h.normalize(); err != nil {
		return Host{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.find(ref)
	if i < 0 {
		return Host{}, ErrNotFound
	}
	prev := s.hosts[i]
	if s.nameTaken(h.Name, prev.ID) {
		return Host{}, fmt.Errorf("%w: a host named %s already exists", ErrConflict, h.Name)
	}
	h.ID = prev.ID
	s.hosts[i] = h
	if err := s.persist(); err != nil {
		s.hosts[i] = prev
		return Host{}, err
	}
	return h, nil
}

func (s *Store) Delete(ref string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.find(ref)
	if i < 0 {
		return ErrNotFound
	}
	prev := slices.Clone(s.hosts)
	s.hosts = slices.Delete(s.hosts, i, i+1)
	if err := s.persist(); err != nil {
		s.hosts = prev
		return err
	}
	return nil
}

func (s *Store) Referencing(keySecretID string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var names []string
	for _, h := range s.hosts {
		if h.KeySecret == keySecretID {
			names = append(names, h.Name)
		}
	}
	slices.Sort(names)
	return names
}

func (s *Store) persist() error {
	data, err := json.Marshal(s.hosts)
	if err != nil {
		return err
	}
	return datadir.WriteFile(s.path, data)
}
