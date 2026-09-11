package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"

	cryptossh "golang.org/x/crypto/ssh"
)

func writeFileExcl(path string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return err
	}
	return nil
}

func writeKeyPair(name string) (string, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}
	sshPub, err := cryptossh.NewPublicKey(pub)
	if err != nil {
		return "", err
	}
	authorized := cryptossh.MarshalAuthorizedKey(sshPub)
	block, err := cryptossh.MarshalPrivateKey(priv, name)
	if err != nil {
		return "", err
	}
	privPEM := pem.EncodeToMemory(block)
	privPath := filepath.Join(dir(), "keys", name)
	pubPath := privPath + ".pub"
	if err := writeFileExcl(privPath, privPEM, 0600); err != nil {
		return "", err
	}
	if err := writeFileExcl(pubPath, authorized, 0600); err != nil {
		os.Remove(privPath)
		return "", err
	}
	return strings.TrimRight(string(authorized), "\n"), nil
}

func removeKeyPair(name string) error {
	base := filepath.Join(dir(), "keys", name)
	for _, p := range []string{base, base + ".pub"} {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}
