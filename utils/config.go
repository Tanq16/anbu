package utils

import (
	"os"
	"path/filepath"
)

func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		PrintFatal("cannot resolve home directory", err)
	}
	dir := filepath.Join(home, ".config", "anbu")
	if err := os.MkdirAll(dir, 0700); err != nil {
		PrintFatal("cannot create config directory", err)
	}
	if err := os.Chmod(dir, 0700); err != nil {
		PrintFatal("cannot set config directory permissions", err)
	}
	return dir
}
