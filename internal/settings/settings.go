package settings

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
	"time"

	"github.com/tanq16/anbu/internal/datadir"
	"github.com/tanq16/anbu/internal/tools"
)

var ErrInvalid = errors.New("invalid settings")

type Settings struct {
	PassphraseWords          int    `json:"passphrase_words"`
	PassphraseSimple         bool   `json:"passphrase_simple"`
	RandomLength             int    `json:"random_length"`
	RandomCharset            string `json:"random_charset"`
	UUIDVersion              int    `json:"uuid_version"`
	AWSCommandTimeoutSeconds int    `json:"aws_command_timeout_seconds"`
	MachineTimezone          string `json:"machine_timezone"`
}

func Defaults() Settings {
	return Settings{
		PassphraseWords:          3,
		PassphraseSimple:         false,
		RandomLength:             48,
		RandomCharset:            "alphanumeric",
		UUIDVersion:              7,
		AWSCommandTimeoutSeconds: 300,
		MachineTimezone:          "",
	}
}

func Decode(data []byte) (Settings, error) {
	s := Defaults()
	if err := json.Unmarshal(data, &s); err != nil {
		return Settings{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	return s, s.Validate()
}

func (s Settings) Validate() error {
	switch {
	case s.PassphraseWords < 1 || s.PassphraseWords > 50:
		return fmt.Errorf("%w: passphrase_words must be between 1 and 50", ErrInvalid)
	case s.RandomLength < 1 || s.RandomLength > 1024:
		return fmt.Errorf("%w: random_length must be between 1 and 1024", ErrInvalid)
	case tools.Charsets[s.RandomCharset] == "":
		return fmt.Errorf("%w: random_charset must be one of alphanumeric, alpha, digits, hex, all", ErrInvalid)
	case s.UUIDVersion != 4 && s.UUIDVersion != 7:
		return fmt.Errorf("%w: uuid_version must be 4 or 7", ErrInvalid)
	case s.AWSCommandTimeoutSeconds < 10 || s.AWSCommandTimeoutSeconds > 3600:
		return fmt.Errorf("%w: aws_command_timeout_seconds must be between 10 and 3600", ErrInvalid)
	}
	if s.MachineTimezone != "" {
		if _, err := time.LoadLocation(s.MachineTimezone); err != nil {
			return fmt.Errorf("%w: machine_timezone %q is not a known time zone", ErrInvalid, s.MachineTimezone)
		}
	}
	return nil
}

type Store struct {
	mu   sync.RWMutex
	path string
	cur  Settings
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, cur: Defaults()}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if s.cur, err = Decode(data); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Get() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cur
}

func (s *Store) Put(next Settings) error {
	if err := next.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(next)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := datadir.WriteFile(s.path, data); err != nil {
		return err
	}
	s.cur = next
	return nil
}
