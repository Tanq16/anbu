package anbuString

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
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
