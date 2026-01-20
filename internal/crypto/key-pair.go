package anbuCrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	u "github.com/tanq16/anbu/utils"
	"golang.org/x/crypto/ssh"
)

func GenerateKeyPair(outputDir, name string, keySize int) {
	u.LineBreak()
	os.MkdirAll(outputDir, 0755)

	privateKey, _ := rsa.GenerateKey(rand.Reader, keySize)
	publicKey := &privateKey.PublicKey
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyPath := fmt.Sprintf("%s/%s.private.pem", outputDir, name)
	privateKeyFile, _ := os.Create(privateKeyPath)
	err := pem.Encode(privateKeyFile, privateKeyBlock)
	privateKeyFile.Close()
	if err != nil {
		u.PrintFatal("failed to write private key to file", err)
	}

	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(publicKey)
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicKeyPath := fmt.Sprintf("%s/%s.public.pem", outputDir, name)
	publicKeyFile, _ := os.Create(publicKeyPath)
	err = pem.Encode(publicKeyFile, publicKeyBlock)
	publicKeyFile.Close()
	if err != nil {
		u.PrintFatal("failed to write public key to file", err)
	}

	u.PrintGeneric(fmt.Sprintf("RSA key pair (%d bits) generated", keySize))
	u.PrintGeneric(fmt.Sprintf("Public key: %s", u.FInfo(publicKeyPath)))
	u.PrintGeneric(fmt.Sprintf("Private key: %s", u.FInfo(privateKeyPath)))
}

func GenerateSSHKeyPair(outputDir, name string, keySize int) {
	u.LineBreak()
	os.MkdirAll(outputDir, 0755)
	privateKey, _ := rsa.GenerateKey(rand.Reader, keySize)
	publicKey, _ := ssh.NewPublicKey(&privateKey.PublicKey)
	sshPublicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)

	publicKeyPath := fmt.Sprintf("%s/%s.pub", outputDir, name)
	if err := os.WriteFile(publicKeyPath, sshPublicKeyBytes, 0644); err != nil {
		u.PrintFatal("failed to write SSH public key file", err)
	}
	privatePEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyPath := fmt.Sprintf("%s/%s", outputDir, name)
	privateKeyFile, _ := os.Create(privateKeyPath)
	err := pem.Encode(privateKeyFile, privatePEM)
	privateKeyFile.Close()
	if err != nil {
		u.PrintFatal("failed to write private key to file", err)
	}
	os.Chmod(privateKeyPath, 0600)

	u.PrintGeneric(fmt.Sprintf("SSH key pair (%d bits) generated", keySize))
	u.PrintGeneric(fmt.Sprintf("Public key: %s", u.FInfo(publicKeyPath)))
	u.PrintGeneric(fmt.Sprintf("Private key: %s", u.FInfo(privateKeyPath)))
}
