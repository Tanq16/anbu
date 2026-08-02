package anbuCrypto

import (
	"os"
	"path/filepath"
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

func TestSymmetricFileEncryption(t *testing.T) {
	tempDir := t.TempDir()
	samplePath := filepath.Join(tempDir, "secret.txt")
	originalContent := "super secret payload data 123"

	if err := os.WriteFile(samplePath, []byte(originalContent), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	password := "correct-horse-battery-staple"

	encPath, err := EncryptFileSymmetric(samplePath, password)
	if err != nil {
		t.Fatalf("EncryptFileSymmetric failed: %v", err)
	}

	if _, err := os.Stat(encPath); err != nil {
		t.Fatalf("encrypted file missing: %v", err)
	}

	decPath, err := DecryptFileSymmetric(encPath, password)
	if err != nil {
		t.Fatalf("DecryptFileSymmetric failed: %v", err)
	}

	decryptedContent, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("failed to read decrypted file: %v", err)
	}

	if string(decryptedContent) != originalContent {
		t.Errorf("expected content %q, got %q", originalContent, string(decryptedContent))
	}

	t.Run("Wrong Password", func(t *testing.T) {
		_, err := DecryptFileSymmetric(encPath, "wrong-password")
		if err == nil {
			t.Error("expected error decrypting with wrong password, got nil")
		}
	})
}
