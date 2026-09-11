package archive

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

var archiveMagic = []byte("ANBUZ")

var archiveSalt = []byte("anbu-archive-v1")

func encryptArchive(data []byte, password string) ([]byte, error) {
	key := pbkdf2.Key([]byte(password), archiveSalt, 100000, 32, sha256.New)
	sealed, err := encryptData(data, key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(archiveMagic)+len(sealed))
	out = append(out, archiveMagic...)
	out = append(out, sealed...)
	return out, nil
}

func decryptArchive(data []byte, password string) ([]byte, error) {
	if !bytes.HasPrefix(data, archiveMagic) {
		return nil, fmt.Errorf("not an encrypted archive")
	}
	key := pbkdf2.Key([]byte(password), archiveSalt, 100000, 32, sha256.New)
	return decryptData(data[len(archiveMagic):], key)
}

func encryptData(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

func decryptData(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
