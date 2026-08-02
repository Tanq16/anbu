package anbuCrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

type KeyPairResult struct {
	KeySize        int
	PublicKeyPath  string
	PrivateKeyPath string
}

func GenerateKeyPair(outputDir, name string, keySize int) (*KeyPairResult, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	publicKey := &privateKey.PublicKey
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyPath := filepath.Join(outputDir, name+".private.pem")
	privateKeyFile, err := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key file: %w", err)
	}
	if err := pem.Encode(privateKeyFile, privateKeyBlock); err != nil {
		privateKeyFile.Close()
		return nil, fmt.Errorf("failed to write private key to file: %w", err)
	}
	privateKeyFile.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicKeyPath := filepath.Join(outputDir, name+".public.pem")
	publicKeyFile, err := os.Create(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create public key file: %w", err)
	}
	if err := pem.Encode(publicKeyFile, publicKeyBlock); err != nil {
		publicKeyFile.Close()
		return nil, fmt.Errorf("failed to write public key to file: %w", err)
	}
	publicKeyFile.Close()

	return &KeyPairResult{
		KeySize:        keySize,
		PublicKeyPath:  publicKeyPath,
		PrivateKeyPath: privateKeyPath,
	}, nil
}

func GenerateSSHKeyPair(outputDir, name string, keySize int) (*KeyPairResult, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH public key: %w", err)
	}
	sshPublicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)

	publicKeyPath := filepath.Join(outputDir, name+".pub")
	if err := os.WriteFile(publicKeyPath, sshPublicKeyBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write SSH public key file: %w", err)
	}
	privatePEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyPath := filepath.Join(outputDir, name)
	privateKeyFile, err := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key file: %w", err)
	}
	if err := pem.Encode(privateKeyFile, privatePEM); err != nil {
		privateKeyFile.Close()
		return nil, fmt.Errorf("failed to write private key to file: %w", err)
	}
	privateKeyFile.Close()

	return &KeyPairResult{
		KeySize:        keySize,
		PublicKeyPath:  publicKeyPath,
		PrivateKeyPath: privateKeyPath,
	}, nil
}
