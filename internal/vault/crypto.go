package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json/v2"
	"errors"
	"fmt"
)

const (
	fileVersion    = 1
	kdfName        = "pbkdf2-sha256"
	kdfIterations  = 600000
	saltSize       = 16
	keySize        = 32
	additionalData = "anbu-vault-v1"
)

type kdf struct {
	Name       string `json:"name"`
	Iterations int    `json:"iterations"`
	Salt       []byte `json:"salt"`
}

type file struct {
	Version int    `json:"version"`
	KDF     kdf    `json:"kdf"`
	Nonce   []byte `json:"nonce"`
	Data    []byte `json:"data"`
}

func deriveKey(password string, salt []byte, iter int) ([]byte, error) {
	return pbkdf2.Key(sha256.New, password, salt, iter, keySize)
}

func newSalt() []byte {
	salt := make([]byte, saltSize)
	rand.Read(salt)
	return salt
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func seal(plain, key, salt []byte, iter int) (file, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return file{}, err
	}
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	return file{
		Version: fileVersion,
		KDF:     kdf{Name: kdfName, Iterations: iter, Salt: salt},
		Nonce:   nonce,
		Data:    gcm.Seal(nil, nonce, plain, []byte(additionalData)),
	}, nil
}

type contents struct {
	Secrets     map[string]Secret     `json:"secrets"`
	MachineKeys map[string]MachineKey `json:"machine_keys"`
}

func sealVault(c contents, key, salt []byte, iter int) ([]byte, error) {
	plain, err := json.Marshal(c, json.Deterministic(true))
	if err != nil {
		return nil, err
	}
	f, err := seal(plain, key, salt, iter)
	if err != nil {
		return nil, err
	}
	return json.Marshal(f)
}

type opened struct {
	contents
	key  []byte
	salt []byte
	iter int
}

func openVault(data []byte, password string) (opened, error) {
	var f file
	if err := json.Unmarshal(data, &f); err != nil {
		return opened{}, fmt.Errorf("vault file is corrupt: %w", err)
	}
	if f.Version != fileVersion || f.KDF.Name != kdfName {
		return opened{}, fmt.Errorf("unsupported vault format version %d with kdf %q", f.Version, f.KDF.Name)
	}
	if f.KDF.Iterations <= 0 || len(f.KDF.Salt) == 0 {
		return opened{}, errors.New("vault file has invalid kdf parameters")
	}
	key, err := deriveKey(password, f.KDF.Salt, f.KDF.Iterations)
	if err != nil {
		return opened{}, err
	}
	gcm, err := newGCM(key)
	if err != nil {
		return opened{}, err
	}
	if len(f.Nonce) != gcm.NonceSize() {
		return opened{}, errors.New("vault file has an invalid nonce")
	}
	plain, err := gcm.Open(nil, f.Nonce, f.Data, []byte(additionalData))
	if err != nil {
		return opened{}, errors.New("vault does not decrypt with the current password")
	}
	c := contents{Secrets: map[string]Secret{}, MachineKeys: map[string]MachineKey{}}
	if err := json.Unmarshal(plain, &c); err != nil {
		return opened{}, fmt.Errorf("vault contents are corrupt: %w", err)
	}
	return opened{contents: c, key: key, salt: f.KDF.Salt, iter: f.KDF.Iterations}, nil
}
