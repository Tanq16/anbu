package anbuKeypair

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func GenerateKeyPair(outputDir, name string, keySize int) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key pair: %w", err)
	}
	publicKey := &privateKey.PublicKey

	// Encode private key to PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// Write private key to file
	privateKeyPath := fmt.Sprintf("%s/%s.private.pem", outputDir, name)
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	err = pem.Encode(privateKeyFile, privateKeyBlock)
	privateKeyFile.Close()
	if err != nil {
		return fmt.Errorf("failed to write private key to file: %w", err)
	}

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	// Write public key to file
	publicKeyPath := fmt.Sprintf("%s/%s.public.pem", outputDir, name)
	publicKeyFile, err := os.Create(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create public key file: %w", err)
	}
	err = pem.Encode(publicKeyFile, publicKeyBlock)
	publicKeyFile.Close()
	if err != nil {
		return fmt.Errorf("failed to write public key to file: %w", err)
	}

	return nil
}
