package anbuFileCrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tanq16/anbu/utils"
)

func EncryptPGP(inputPath, recipientPubKeyPath, signerPrivKeyPath, passphrase string) error {
	logger := utils.GetLogger("filecrypto")

	// Read the file content
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Generate a random session key (AES key)
	sessionKey := make([]byte, 32) // 256-bit key for AES-256
	if _, err := io.ReadFull(rand.Reader, sessionKey); err != nil {
		return fmt.Errorf("failed to generate session key: %w", err)
	}

	// Encrypt the file with the session key using AES-GCM
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM mode: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	encryptedContent := gcm.Seal(nonce, nonce, content, nil)

	// Read recipient's public key
	recipientPubKey, err := readPublicKey(recipientPubKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read recipient's public key: %w", err)
	}

	// Encrypt the session key with recipient's public key
	encryptedSessionKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		recipientPubKey,
		sessionKey,
		[]byte("session-key"),
	)
	if err != nil {
		return fmt.Errorf("failed to encrypt session key: %w", err)
	}

	// If a signer private key path is provided, create a signature
	var signature []byte
	if signerPrivKeyPath != "" {
		// Read signer's private key
		signerPrivKey, err := readPrivateKey(signerPrivKeyPath, passphrase)
		if err != nil {
			return fmt.Errorf("failed to read signer's private key: %w", err)
		}

		// Create a signature for the encrypted content
		hashed := sha256.Sum256(encryptedContent)
		signature, err = rsa.SignPKCS1v15(rand.Reader, signerPrivKey, 0, hashed[:])
		if err != nil {
			return fmt.Errorf("failed to sign data: %w", err)
		}
	}

	// Format the output packet
	packet := PGPPacket{
		EncryptedSessionKey: encryptedSessionKey,
		EncryptedContent:    encryptedContent,
		Signature:           signature,
		HasSignature:        len(signature) > 0,
	}

	// Serialize the packet
	serialized, err := serializePacket(packet)
	if err != nil {
		return fmt.Errorf("failed to serialize PGP packet: %w", err)
	}

	// Write to output file
	outputPath := inputPath + ".pgp"
	err = os.WriteFile(outputPath, serialized, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	logger.Debug().
		Str("input", inputPath).
		Str("output", outputPath).
		Bool("signed", len(signature) > 0).
		Msg("file encrypted with PGP-like encryption")

	return nil
}

func DecryptPGP(inputPath, recipientPrivKeyPath, signerPubKeyPath, passphrase string) error {
	logger := utils.GetLogger("filecrypto")

	// Read the encrypted file
	encryptedData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Deserialize the packet
	packet, err := deserializePacket(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to deserialize PGP packet: %w", err)
	}

	// Read recipient's private key
	recipientPrivKey, err := readPrivateKey(recipientPrivKeyPath, passphrase)
	if err != nil {
		return fmt.Errorf("failed to read recipient's private key: %w", err)
	}

	// Decrypt the session key
	sessionKey, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		recipientPrivKey,
		packet.EncryptedSessionKey,
		[]byte("session-key"),
	)
	if err != nil {
		return fmt.Errorf("failed to decrypt session key: %w", err)
	}

	// If there's a signature and signer's public key is provided, verify the signature
	if packet.HasSignature && signerPubKeyPath != "" {
		// Read signer's public key
		signerPubKey, err := readPublicKey(signerPubKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read signer's public key: %w", err)
		}

		// Verify the signature
		hashed := sha256.Sum256(packet.EncryptedContent)
		err = rsa.VerifyPKCS1v15(signerPubKey, 0, hashed[:], packet.Signature)
		if err != nil {
			return fmt.Errorf("signature verification failed: %w", err)
		}
		logger.Debug().Msg("signature verified successfully")
	}

	// Decrypt the content with the session key
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(packet.EncryptedContent) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := packet.EncryptedContent[:nonceSize], packet.EncryptedContent[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt content: %w", err)
	}

	// Write decrypted data to output file
	outputPath := strings.TrimSuffix(inputPath, ".pgp")
	err = os.WriteFile(outputPath, plaintext, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	logger.Debug().
		Str("input", inputPath).
		Str("output", outputPath).
		Bool("signature_verified", packet.HasSignature && signerPubKeyPath != "").
		Msg("file decrypted with PGP-like decryption")

	return nil
}

// Helper types and functions

type PGPPacket struct {
	EncryptedSessionKey []byte
	EncryptedContent    []byte
	Signature           []byte
	HasSignature        bool
}

func serializePacket(packet PGPPacket) ([]byte, error) {
	// Format:
	// - 8 bytes: Length of encrypted session key
	// - N bytes: Encrypted session key
	// - 8 bytes: Length of encrypted content
	// - M bytes: Encrypted content
	// - 1 byte: Has signature (0 or 1)
	// - 8 bytes: Length of signature (if has signature)
	// - P bytes: Signature (if has signature)

	// Calculate total size
	totalSize := 8 + len(packet.EncryptedSessionKey) + 8 + len(packet.EncryptedContent) + 1
	if packet.HasSignature {
		totalSize += 8 + len(packet.Signature)
	}

	// Create buffer
	buf := make([]byte, totalSize)
	offset := 0

	// Write session key length and data
	copy(buf[offset:offset+8], fmt.Sprintf("%08d", len(packet.EncryptedSessionKey)))
	offset += 8
	copy(buf[offset:offset+len(packet.EncryptedSessionKey)], packet.EncryptedSessionKey)
	offset += len(packet.EncryptedSessionKey)

	// Write encrypted content length and data
	copy(buf[offset:offset+8], fmt.Sprintf("%08d", len(packet.EncryptedContent)))
	offset += 8
	copy(buf[offset:offset+len(packet.EncryptedContent)], packet.EncryptedContent)
	offset += len(packet.EncryptedContent)

	// Write signature flag
	if packet.HasSignature {
		buf[offset] = 1
	} else {
		buf[offset] = 0
	}
	offset++

	// Write signature length and data if present
	if packet.HasSignature {
		copy(buf[offset:offset+8], fmt.Sprintf("%08d", len(packet.Signature)))
		offset += 8
		copy(buf[offset:offset+len(packet.Signature)], packet.Signature)
	}

	// Encode the entire buffer in base64
	encoded := base64.StdEncoding.EncodeToString(buf)
	return []byte(encoded), nil
}

func deserializePacket(data []byte) (PGPPacket, error) {
	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return PGPPacket{}, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Parse the packet
	offset := 0

	// Check if there's enough data for session key length
	if len(decoded) < offset+8 {
		return PGPPacket{}, fmt.Errorf("invalid packet format: too short for session key length")
	}

	// Extract session key length
	sessionKeyLenStr := string(decoded[offset : offset+8])
	sessionKeyLen, err := fmt.Sscanf(sessionKeyLenStr, "%d", new(int))
	if err != nil || sessionKeyLen != 1 {
		return PGPPacket{}, fmt.Errorf("invalid packet format: invalid session key length")
	}
	sessionKeyLen = int(sessionKeyLenStr[0]-'0')*10000000 +
		int(sessionKeyLenStr[1]-'0')*1000000 +
		int(sessionKeyLenStr[2]-'0')*100000 +
		int(sessionKeyLenStr[3]-'0')*10000 +
		int(sessionKeyLenStr[4]-'0')*1000 +
		int(sessionKeyLenStr[5]-'0')*100 +
		int(sessionKeyLenStr[6]-'0')*10 +
		int(sessionKeyLenStr[7]-'0')
	offset += 8

	// Check if there's enough data for session key
	if len(decoded) < offset+sessionKeyLen {
		return PGPPacket{}, fmt.Errorf("invalid packet format: too short for session key")
	}

	// Extract session key
	encryptedSessionKey := decoded[offset : offset+sessionKeyLen]
	offset += sessionKeyLen

	// Check if there's enough data for content length
	if len(decoded) < offset+8 {
		return PGPPacket{}, fmt.Errorf("invalid packet format: too short for content length")
	}

	// Extract content length
	contentLenStr := string(decoded[offset : offset+8])
	contentLen, err := fmt.Sscanf(contentLenStr, "%d", new(int))
	if err != nil || contentLen != 1 {
		return PGPPacket{}, fmt.Errorf("invalid packet format: invalid content length")
	}
	contentLen = int(contentLenStr[0]-'0')*10000000 +
		int(contentLenStr[1]-'0')*1000000 +
		int(contentLenStr[2]-'0')*100000 +
		int(contentLenStr[3]-'0')*10000 +
		int(contentLenStr[4]-'0')*1000 +
		int(contentLenStr[5]-'0')*100 +
		int(contentLenStr[6]-'0')*10 +
		int(contentLenStr[7]-'0')
	offset += 8

	// Check if there's enough data for content
	if len(decoded) < offset+contentLen {
		return PGPPacket{}, fmt.Errorf("invalid packet format: too short for content")
	}

	// Extract content
	encryptedContent := decoded[offset : offset+contentLen]
	offset += contentLen

	// Check if there's enough data for signature flag
	if len(decoded) < offset+1 {
		return PGPPacket{}, fmt.Errorf("invalid packet format: too short for signature flag")
	}

	// Extract signature flag
	hasSignature := decoded[offset] == 1
	offset++

	// If there's a signature, extract it
	var signature []byte
	if hasSignature {
		// Check if there's enough data for signature length
		if len(decoded) < offset+8 {
			return PGPPacket{}, fmt.Errorf("invalid packet format: too short for signature length")
		}

		// Extract signature length
		sigLenStr := string(decoded[offset : offset+8])
		sigLen, err := fmt.Sscanf(sigLenStr, "%d", new(int))
		if err != nil || sigLen != 1 {
			return PGPPacket{}, fmt.Errorf("invalid packet format: invalid signature length")
		}
		sigLen = int(sigLenStr[0]-'0')*10000000 +
			int(sigLenStr[1]-'0')*1000000 +
			int(sigLenStr[2]-'0')*100000 +
			int(sigLenStr[3]-'0')*10000 +
			int(sigLenStr[4]-'0')*1000 +
			int(sigLenStr[5]-'0')*100 +
			int(sigLenStr[6]-'0')*10 +
			int(sigLenStr[7]-'0')
		offset += 8

		// Check if there's enough data for signature
		if len(decoded) < offset+sigLen {
			return PGPPacket{}, fmt.Errorf("invalid packet format: too short for signature")
		}

		// Extract signature
		signature = decoded[offset : offset+sigLen]
	}

	return PGPPacket{
		EncryptedSessionKey: encryptedSessionKey,
		EncryptedContent:    encryptedContent,
		Signature:           signature,
		HasSignature:        hasSignature,
	}, nil
}

// Key management functions

func readPublicKey(path string) (*rsa.PublicKey, error) {
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}

func readPrivateKey(path, passphrase string) (*rsa.PrivateKey, error) {
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	var keyBytes []byte
	if x509.IsEncryptedPEMBlock(block) {
		keyBytes, err = x509.DecryptPEMBlock(block, []byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt private key: %w", err)
		}
	} else {
		keyBytes = block.Bytes
	}

	var parsedKey any
	if block.Type == "RSA PRIVATE KEY" {
		parsedKey, err = x509.ParsePKCS1PrivateKey(keyBytes)
	} else {
		parsedKey, err = x509.ParsePKCS8PrivateKey(keyBytes)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	privateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return privateKey, nil
}

// Derive a 32-byte key from a password using SHA-256
func deriveKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}
