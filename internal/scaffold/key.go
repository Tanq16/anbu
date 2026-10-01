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

var ErrKeyMismatch = errors.New("private key does not match the key pair")

type KeyStatus struct {
	State       KeyState `json:"state"`
	Fingerprint string   `json:"fingerprint,omitempty"`
	keyPair     *awsx.KeyPair
	stored      *vault.MachineKey
}

func CheckKey(ctx context.Context, c *awsx.Clients, keys *vault.Store) (KeyStatus, error) {
	kp, err := c.FindKeyPair(ctx)
	if err != nil {
		return KeyStatus{}, err
	}
	status := KeyStatus{keyPair: kp}
	if k, ok := keys.MachineKey(c.Account, c.Region); ok {
		status.stored = &k
		status.Fingerprint = k.Fingerprint()
	}

	switch {
	case kp == nil:
		status.State = KeyNone
	case status.stored == nil:
		status.State = KeyMissing
	case samePublicKey(kp.PublicKey, status.stored.PublicKey):
		status.State = KeyMatch
	default:
		status.State = KeyDiffers
	}
	return status, nil
}

func AdoptKey(ctx context.Context, c *awsx.Clients, keys *vault.Store, privateKey string) (KeyStatus, error) {
	kp, err := c.FindKeyPair(ctx)
	if err != nil {
		return KeyStatus{}, err
	}
	if kp != nil {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return KeyStatus{}, fmt.Errorf("%w: private key does not parse: %v", vault.ErrInvalid, err)
		}
		if !samePublicKey(kp.PublicKey, string(ssh.MarshalAuthorizedKey(signer.PublicKey()))) {
			return KeyStatus{}, fmt.Errorf("%w %s in %s", ErrKeyMismatch, nameKeyPair, c.Region)
		}
	}
	if _, err := keys.ImportMachineKey(c.Account, c.Region, privateKey); err != nil {
		return KeyStatus{}, err
	}
	return CheckKey(ctx, c, keys)
}

func (k KeyStatus) Err(region string) error {
	switch k.State {
	case KeyMissing:
		return fmt.Errorf("key pair %s exists in %s but anbu has no key for it; import its private key on the scaffold view", nameKeyPair, region)
	case KeyDiffers:
		return fmt.Errorf("key pair %s in %s does not match the scaffold key; machines launched there authorize the other key", nameKeyPair, region)
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
