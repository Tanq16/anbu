package anbuGenerics

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		length = 100
	}
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(randomBytes)
	encoded = strings.Map(func(r rune) rune {
		switch r {
		case '=', '+', '/', '\n':
			return -1
		default:
			return r
		}
	}, encoded)
	if len(encoded) > length {
		encoded = encoded[:length]
	}
	return encoded, nil
}

// generates a sequence of alphabetic characters
func GenerateSequence(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than 0")
	}
	alphabet := "abcdefghijklmnopqrstuvxyz"
	var result strings.Builder
	for result.Len() < length {
		result.WriteString(alphabet)
	}
	return result.String()[:length], nil
}

// repeats a string a specified number of times
func GenerateRepetition(count int, str string) (string, error) {
	if count <= 0 {
		return "", fmt.Errorf("count must be greater than 0")
	}
	var result strings.Builder
	for range count {
		result.WriteString(str)
	}
	return result.String(), nil
}

// generates a UUID string
func GenerateUUID() (string, error) {
	// use google/uuid package to generate a UUID
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return uuid.String(), nil
}

// generates shorter UUID string
func GenerateRUID(len string) (string, error) {
	length, err := strconv.Atoi(len)
	if err != nil {
		return "", fmt.Errorf("not a valid length: %w", err)
	}
	if length <= 0 || length > 32 {
		return "", fmt.Errorf("length must be between 1 and 30")
	}
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	// remove version and variant bits from UUID
	shortUUID := uuid.String()[0:8] + uuid.String()[9:13] + uuid.String()[15:18] + uuid.String()[20:23] + uuid.String()[24:]
	// shortUUID := strings.ReplaceAll(uuid.String(), "-", "")[:length]
	return shortUUID[:length], nil
}
