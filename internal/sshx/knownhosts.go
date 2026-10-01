package sshx

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/tanq16/anbu/internal/datadir"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type KnownHosts struct {
	mu   sync.Mutex
	path string
}

func NewKnownHosts(path string) *KnownHosts {
	return &KnownHosts{path: path}
}

func (k *KnownHosts) Callback(alias string) ssh.HostKeyCallback {
	return func(_ string, remote net.Addr, key ssh.PublicKey) error {
		k.mu.Lock()
		defer k.mu.Unlock()
		check, err := knownhosts.New(k.path)
		if err != nil {
			return err
		}
		// knownhosts rejects an address without a port, and a port-22 alias in known_hosts form has none.
		err = check(withPort(alias), remote, key)
		keyErr, ok := errors.AsType[*knownhosts.KeyError](err)
		if !ok || len(keyErr.Want) > 0 {
			return err
		}
		f, err := os.OpenFile(k.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(knownhosts.Line([]string{alias}, key) + "\n")
		return err
	}
}

func withPort(alias string) string {
	if _, _, err := net.SplitHostPort(alias); err == nil {
		return alias
	}
	return net.JoinHostPort(alias, "22")
}

func (k *KnownHosts) Remove(alias string) (int, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	data, err := os.ReadFile(k.path)
	if err != nil {
		return 0, err
	}
	want := knownhosts.Normalize(alias)
	var kept bytes.Buffer
	removed := 0
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		if fields := strings.Fields(line); len(fields) > 0 && !strings.HasPrefix(fields[0], "#") &&
			slices.Contains(strings.Split(fields[0], ","), want) {
			removed++
			continue
		}
		kept.WriteString(line + "\n")
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	if removed == 0 {
		return 0, nil
	}
	return removed, datadir.WriteFile(k.path, kept.Bytes())
}
