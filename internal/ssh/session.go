package ssh

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/tanq16/anbu/utils"
)

type Session struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Port     int    `json:"port"`
	Identity string `json:"identity"`
}

type SetupConfig struct {
	Name string
	Host string
	User string
	Port int
}

func dir() string {
	return filepath.Join(utils.ConfigDir(), "ssh")
}

func sessionsPath() string {
	return filepath.Join(dir(), "sessions.json")
}

func knownHostsPath() string {
	return filepath.Join(dir(), "known_hosts")
}

func identityFile(identity string) string {
	if filepath.IsAbs(identity) {
		return identity
	}
	return filepath.Join(dir(), identity)
}

func validName(name string) error {
	if name == "" || name != filepath.Base(name) || name == "." || name == ".." {
		return fmt.Errorf("invalid session name %q", name)
	}
	return nil
}

func ensureDir() error {
	keys := filepath.Join(dir(), "keys")
	if err := os.MkdirAll(keys, 0700); err != nil {
		return err
	}
	if err := os.Chmod(dir(), 0700); err != nil {
		return err
	}
	if err := os.Chmod(keys, 0700); err != nil {
		return err
	}
	kh := knownHostsPath()
	_, err := os.Stat(kh)
	if errors.Is(err, os.ErrNotExist) {
		return os.WriteFile(kh, []byte{}, 0600)
	}
	return err
}

func loadStore() (map[string]Session, error) {
	data, err := os.ReadFile(sessionsPath())
	if errors.Is(err, os.ErrNotExist) {
		return map[string]Session{}, nil
	}
	if err != nil {
		return nil, err
	}
	var store map[string]Session
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	if store == nil {
		store = map[string]Session{}
	}
	return store, nil
}

func saveStore(store map[string]Session) error {
	data, err := json.Marshal(store, jsontext.WithIndent("  "))
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(sessionsPath(), data, 0600)
}

func get(name string) (Session, error) {
	store, err := loadStore()
	if err != nil {
		return Session{}, err
	}
	sess, ok := store[name]
	if !ok {
		return Session{}, fmt.Errorf("session %q not found", name)
	}
	sess.Name = name
	return sess, nil
}

func List() ([]Session, error) {
	store, err := loadStore()
	if err != nil {
		return nil, err
	}
	names := slices.Sorted(maps.Keys(store))
	out := make([]Session, 0, len(names))
	for _, name := range names {
		s := store[name]
		s.Name = name
		out = append(out, s)
	}
	return out, nil
}

func Setup(cfg SetupConfig) (string, error) {
	if err := validName(cfg.Name); err != nil {
		return "", err
	}
	if cfg.Host == "" {
		return "", fmt.Errorf("host is required")
	}
	if cfg.User == "" {
		return "", fmt.Errorf("user is required")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return "", fmt.Errorf("invalid port %d", cfg.Port)
	}
	if err := ensureDir(); err != nil {
		return "", err
	}
	store, err := loadStore()
	if err != nil {
		return "", err
	}
	if _, exists := store[cfg.Name]; exists {
		return "", fmt.Errorf("session %q already exists", cfg.Name)
	}
	pub, err := writeKeyPair(cfg.Name)
	if err != nil {
		return "", err
	}
	store[cfg.Name] = Session{
		Name:     cfg.Name,
		Host:     cfg.Host,
		User:     cfg.User,
		Port:     cfg.Port,
		Identity: filepath.Join("keys", cfg.Name),
	}
	if err := saveStore(store); err != nil {
		_ = removeKeyPair(cfg.Name)
		return "", err
	}
	return pub, nil
}

func Delete(name string) error {
	if err := validName(name); err != nil {
		return err
	}
	store, err := loadStore()
	if err != nil {
		return err
	}
	if _, ok := store[name]; !ok {
		return fmt.Errorf("session %q not found", name)
	}
	if err := removeKeyPair(name); err != nil {
		return err
	}
	delete(store, name)
	return saveStore(store)
}
