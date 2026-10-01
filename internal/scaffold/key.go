package scaffold

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/tanq16/anbu/internal/awsx"
	"github.com/tanq16/anbu/internal/vault"
	"golang.org/x/crypto/ssh"
)

type KeyState string

const (
	KeyNone    KeyState = "none"
	KeyMatch   KeyState = "match"
	KeyMissing KeyState = "missing"
	KeyDiffers KeyState = "differs"
)

type KeyStatus struct {
	Secret  string   `json:"secret"`
	State   KeyState `json:"state"`
	keyPair *awsx.KeyPair
	stored  *vault.Secret
}

func CheckKey(ctx context.Context, c *awsx.Clients, keys *vault.Store) (KeyStatus, error) {
	status := KeyStatus{Secret: KeyName(c.Account, c.Region)}
	kp, err := c.FindKeyPair(ctx)
	if err != nil {
		return KeyStatus{}, err
	}
	status.keyPair = kp
	sec, err := keys.Get(status.Secret)
	switch {
	case errors.Is(err, vault.ErrNotFound):
	case err != nil:
		return KeyStatus{}, err
	default:
		status.stored = &sec
	}

	switch {
	case kp == nil:
		status.State = KeyNone
	case status.stored == nil:
		status.State = KeyMissing
	case samePublicKey(kp.PublicKey, status.stored.Fields["public_key"]):
		status.State = KeyMatch
	default:
		status.State = KeyDiffers
	}
	return status, nil
}

func (k KeyStatus) Err(region string) error {
	switch k.State {
	case KeyMissing:
		return fmt.Errorf("key pair %s exists in %s but the vault has no secret %s; import its private key as an ssh-key secret under that name",
			nameKeyPair, region, k.Secret)
	case KeyDiffers:
		return fmt.Errorf("key pair %s in %s does not match the secret %s; machines launched there authorize the other key",
			nameKeyPair, region, k.Secret)
	}
	return nil
}

func samePublicKey(a, b string) bool {
	ka, _, _, _, err := ssh.ParseAuthorizedKey([]byte(a))
	if err != nil {
		return false
	}
	kb, _, _, _, err := ssh.ParseAuthorizedKey([]byte(b))
	if err != nil {
		return false
	}
	return bytes.Equal(ka.Marshal(), kb.Marshal())
}

func (k KeyStatus) KeyPairName() string {
	if k.keyPair == nil {
		return ""
	}
	return k.keyPair.Name
}
