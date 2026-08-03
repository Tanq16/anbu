package anbuCrypto

import (
	"os"
	"testing"
)

func TestKeyPairGeneration(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("RSA Key Pair", func(t *testing.T) {
		res, err := GenerateKeyPair(tempDir, "testrsa", 2048)
		if err != nil {
			t.Fatalf("GenerateKeyPair failed: %v", err)
		}
		if res.KeySize != 2048 {
			t.Errorf("expected KeySize 2048, got %d", res.KeySize)
		}
		if _, err := os.Stat(res.PrivateKeyPath); err != nil {
			t.Errorf("private key file missing: %v", err)
		}
		if _, err := os.Stat(res.PublicKeyPath); err != nil {
			t.Errorf("public key file missing: %v", err)
		}

		info, err := os.Stat(res.PrivateKeyPath)
		if err != nil {
			t.Fatalf("failed to stat private key: %v", err)
		}
		if mode := info.Mode().Perm(); mode != 0600 {
			t.Errorf("expected mode 0600, got %o", mode)
		}
	})

	t.Run("SSH Key Pair", func(t *testing.T) {
		res, err := GenerateSSHKeyPair(tempDir, "testssh", 2048)
		if err != nil {
			t.Fatalf("GenerateSSHKeyPair failed: %v", err)
		}
		if _, err := os.Stat(res.PrivateKeyPath); err != nil {
			t.Errorf("SSH private key missing: %v", err)
		}
		if _, err := os.Stat(res.PublicKeyPath); err != nil {
			t.Errorf("SSH public key missing: %v", err)
		}
	})
}
