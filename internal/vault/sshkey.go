package vault

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

func Signer(s Secret) (ssh.Signer, error) {
	if s.Type != TypeSSHKey {
		return nil, fmt.Errorf("secret %s has type %s, not %s", s.Name, s.Type, TypeSSHKey)
	}
	key := []byte(s.Fields["private_key"])
	if passphrase := s.Fields["passphrase"]; passphrase != "" {
		return ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	}
	return ssh.ParsePrivateKey(key)
}

func (s *Secret) fillPublicKey() error {
	signer, err := Signer(*s)
	if err != nil {
		return invalid("private_key does not parse: %v", err)
	}
	if s.Fields["public_key"] == "" {
		s.Fields["public_key"] = string(ssh.MarshalAuthorizedKey(signer.PublicKey()))
	}
	return nil
}

func newSSHKeyFields(comment string) (map[string]string, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	block, err := ssh.MarshalPrivateKey(priv, comment)
	if err != nil {
		return nil, err
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"private_key": string(pem.EncodeToMemory(block)),
		"public_key":  string(ssh.MarshalAuthorizedKey(sshPub)),
		"passphrase":  "",
	}, nil
}
