package anbuGenerics

import (
	cryptoRand "crypto/rand"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

const randomCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		length = 100
	}
	var sb strings.Builder
	sb.Grow(length)
	charsetLen := big.NewInt(int64(len(randomCharset)))
	for range length {
		idx, err := cryptoRand.Int(cryptoRand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		sb.WriteByte(randomCharset[idx.Int64()])
	}
	return sb.String(), nil
}

func GenerateSequenceString(length int) string {
	if length <= 0 {
		length = 100
	}
	alphabet := "abcdefghijklmnopqrstuvwxyz"
	var result strings.Builder
	for result.Len() < length {
		result.WriteString(alphabet)
	}
	return result.String()[:length]
}

func GenerateRepetitionString(count int, str string) string {
	if count <= 0 {
		count = 10
	}
	var result strings.Builder
	for range count {
		result.WriteString(str)
	}
	return result.String()
}

func GenerateUUIDString() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func GenerateRUIDString(length int) (string, error) {
	if length <= 0 || length > 32 {
		length = 18
	}
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	cleanUUID := strings.ReplaceAll(u.String(), "-", "")
	return cleanUUID[:length], nil
}

