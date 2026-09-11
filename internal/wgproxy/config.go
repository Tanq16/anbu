package wgproxy

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"
)

type Config struct {
	PrivateKey    string
	PeerPublicKey string
	PresharedKey  string
	Endpoint      string
	Address       string
	DNS           string
	ListenAddr    string
	MTU           int
	KeepAlive     int
	DialTimeout   time.Duration
	Debug         bool
}

func NormalizeKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("key is empty")
	}
	if len(key) == 64 {
		decoded, err := hex.DecodeString(key)
		if err == nil && len(decoded) == 32 {
			return strings.ToLower(key), nil
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(key)
	}
	if err != nil {
		return "", err
	}
	if len(decoded) != 32 {
		return "", fmt.Errorf("invalid key length: expected 32 bytes, got %d", len(decoded))
	}
	return hex.EncodeToString(decoded), nil
}

func parseTunnelAddr(value string) (netip.Addr, error) {
	value = strings.TrimSpace(value)
	if prefix, err := netip.ParsePrefix(value); err == nil {
		return prefix.Addr(), nil
	}
	return netip.ParseAddr(value)
}

func (c *Config) applyDefaults() {
	if c.DNS == "" {
		c.DNS = "1.1.1.1"
	}
	if c.MTU <= 0 {
		c.MTU = 1420
	}
	if c.ListenAddr == "" {
		c.ListenAddr = "127.0.0.1:8888"
	}
}

func (c Config) validate() error {
	if c.PrivateKey == "" {
		return errors.New("private key is required")
	}
	if c.PeerPublicKey == "" {
		return errors.New("peer public key is required")
	}
	if c.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	if c.Address == "" {
		return errors.New("address is required")
	}
	if c.KeepAlive < 0 {
		return errors.New("keepalive must be >= 0")
	}
	if c.DialTimeout < 0 {
		return errors.New("timeout must be >= 0")
	}
	if _, err := parseTunnelAddr(c.Address); err != nil {
		return fmt.Errorf("invalid tunnel address: %w", err)
	}
	if _, err := netip.ParseAddr(c.DNS); err != nil {
		return fmt.Errorf("invalid DNS address: %w", err)
	}
	return nil
}
