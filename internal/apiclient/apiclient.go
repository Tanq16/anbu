package apiclient

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tanq16/anbu/internal/datadir"
)

const (
	DefaultURL = "http://localhost:8080"
	envURL     = "ANBU_URL"
)

type Config struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitzero"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "anbu"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "api.json"), nil
}

func Load() (Config, error) {
	cfg := Config{URL: DefaultURL}
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return Config{}, err
	default:
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("%s: %w", path, err)
		}
	}
	if env := os.Getenv(envURL); env != "" {
		cfg.URL = env
	}
	return cfg, nil
}

func Save(cfg Config) (string, error) {
	if err := validURL(cfg.URL); err != nil {
		return "", err
	}
	path, err := configPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	data, err := json.Marshal(cfg, jsontext.WithIndent("  "))
	if err != nil {
		return "", err
	}
	return path, datadir.WriteFile(path, append(data, '\n'))
}

func ParseHeader(h string) (string, string, error) {
	key, value, ok := strings.Cut(h, ":")
	key, value = strings.TrimSpace(key), strings.TrimSpace(value)
	if !ok || key == "" {
		return "", "", fmt.Errorf("header %q is not in Key: Value form", h)
	}
	return key, value, nil
}

func validURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("url %q must be http or https with a host", raw)
	}
	return nil
}

func newClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 5 * time.Minute,
			MaxIdleConns:          10,
			IdleConnTimeout:       90 * time.Second,
		},
	}
}

func Do(cfg Config, method, path string, body any) (*http.Response, error) {
	if err := validURL(cfg.URL); err != nil {
		return nil, err
	}
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, strings.TrimRight(cfg.URL, "/")+path, reader)
	if err != nil {
		return nil, err
	}
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return newClient().Do(req)
}
