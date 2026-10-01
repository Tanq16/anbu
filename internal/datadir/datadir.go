package datadir

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tanq16/anbu/internal/tools"
)

const (
	PasswordFile   = "password"
	VaultFile      = "vault.json"
	TasksFile      = "tasks.json"
	SettingsFile   = "settings.json"
	HostsFile      = "hosts.json"
	KnownHostsFile = "known_hosts"
	AWSDir         = "aws"
	AWSConfig      = "aws/config"
	AWSCredentials = "aws/credentials"
	AWSHome        = "aws/home"
)

func Init(root string) (string, error) {
	if root == "" {
		return "", errors.New("data directory is required")
	}
	for _, dir := range []string{root, filepath.Join(root, AWSDir), filepath.Join(root, AWSHome)} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return "", err
		}
		if err := os.Chmod(dir, 0o700); err != nil {
			return "", err
		}
	}
	for _, name := range []string{AWSConfig, AWSCredentials, KnownHostsFile} {
		if err := ensureFile(filepath.Join(root, name)); err != nil {
			return "", err
		}
	}
	return ensurePassword(filepath.Join(root, PasswordFile))
}

func ReadPassword(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, PasswordFile))
	if err != nil {
		return "", err
	}
	password := strings.TrimRight(string(data), "\r\n")
	if password == "" {
		return "", errors.New("password file is empty")
	}
	return password, nil
}

func ensurePassword(path string) (string, error) {
	_, err := os.Stat(path)
	if err == nil {
		return ReadPassword(filepath.Dir(path))
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	password, err := tools.GenerateRandomStringCharset(32, tools.CharsetAlphaNum)
	if err != nil {
		return "", err
	}
	if err := WriteFile(path, []byte(password+"\n")); err != nil {
		return "", err
	}
	return password, nil
}

func ensureFile(path string) error {
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return WriteFile(path, nil)
	}
	return err
}
