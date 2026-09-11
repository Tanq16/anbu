package wgproxy

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Load(filePath string, flags Config, changed func(string) bool) (Config, error) {
	cfg := flags
	if filePath == "" {
		return cfg, nil
	}
	file, err := parseWGQuick(filePath)
	if err != nil {
		return Config{}, err
	}
	if changed == nil {
		changed = func(string) bool { return false }
	}
	if !changed("private-key") && file.cfg.PrivateKey != "" {
		cfg.PrivateKey = file.cfg.PrivateKey
	}
	if !changed("peer-key") && file.cfg.PeerPublicKey != "" {
		cfg.PeerPublicKey = file.cfg.PeerPublicKey
	}
	if !changed("preshared-key") && file.cfg.PresharedKey != "" {
		cfg.PresharedKey = file.cfg.PresharedKey
	}
	if !changed("endpoint") && file.cfg.Endpoint != "" {
		cfg.Endpoint = file.cfg.Endpoint
	}
	if !changed("address") && file.cfg.Address != "" {
		cfg.Address = file.cfg.Address
	}
	if !changed("dns") && file.cfg.DNS != "" {
		cfg.DNS = file.cfg.DNS
	}
	if !changed("mtu") && file.mtuSet {
		cfg.MTU = file.cfg.MTU
	}
	if !changed("keepalive") && file.keepAliveSet {
		cfg.KeepAlive = file.cfg.KeepAlive
	}
	return cfg, nil
}

type parsedFile struct {
	cfg          Config
	mtuSet       bool
	keepAliveSet bool
}

func parseWGQuick(path string) (parsedFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return parsedFile{}, err
	}

	var out parsedFile
	section := ""
	peers := 0
	for line := range strings.SplitSeq(string(data), "\n") {
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			if section == "peer" {
				peers++
				if peers > 1 {
					return parsedFile{}, errors.New("wg-proxy needs a single [Peer] section")
				}
			}
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			return parsedFile{}, fmt.Errorf("invalid config line: %s", line)
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		switch section {
		case "interface":
			if err := applyInterfaceField(&out, key, value); err != nil {
				return parsedFile{}, err
			}
		case "peer":
			if err := applyPeerField(&out, key, value); err != nil {
				return parsedFile{}, err
			}
		}
	}
	if out.cfg.PrivateKey == "" || out.cfg.PeerPublicKey == "" || out.cfg.Endpoint == "" || out.cfg.Address == "" {
		return parsedFile{}, errors.New("config file needs Interface Address, Interface PrivateKey, Peer PublicKey, and Peer Endpoint")
	}
	return out, nil
}

func applyInterfaceField(out *parsedFile, key, value string) error {
	switch key {
	case "privatekey":
		out.cfg.PrivateKey = value
	case "address":
		out.cfg.Address = firstCSV(value)
	case "dns":
		out.cfg.DNS = firstCSV(value)
	case "mtu":
		mtu, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid MTU: %w", err)
		}
		out.cfg.MTU = mtu
		out.mtuSet = true
	}
	return nil
}

func applyPeerField(out *parsedFile, key, value string) error {
	switch key {
	case "publickey":
		out.cfg.PeerPublicKey = value
	case "presharedkey":
		out.cfg.PresharedKey = value
	case "endpoint":
		out.cfg.Endpoint = value
	case "persistentkeepalive":
		keepalive, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid PersistentKeepalive: %w", err)
		}
		out.cfg.KeepAlive = keepalive
		out.keepAliveSet = true
	}
	return nil
}

func firstCSV(value string) string {
	for part := range strings.SplitSeq(value, ",") {
		if s := strings.TrimSpace(part); s != "" {
			return s
		}
	}
	return ""
}
