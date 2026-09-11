package anbuGenerics

import (
	cryptoRand "crypto/rand"
	"math/big"
	"strings"
	"uuid"
)

const (
	CharsetAlphaNum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	CharsetAlpha    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetDigits   = "0123456789"
	CharsetHex      = "0123456789abcdef"
	CharsetAll      = CharsetAlphaNum + "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
)

func GenerateRandomString(length int) (string, error) {
	return GenerateRandomStringCharset(length, CharsetAlphaNum)
}

func GenerateRandomStringCharset(length int, charset string) (string, error) {
	if length <= 0 {
		length = 100
	}
	if charset == "" {
		charset = CharsetAlphaNum
	}
	var sb strings.Builder
	sb.Grow(length)
	charsetLen := big.NewInt(int64(len(charset)))
	for range length {
		idx, err := cryptoRand.Int(cryptoRand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		sb.WriteByte(charset[idx.Int64()])
	}
	return sb.String(), nil
}

func GenerateUUIDString(v4 bool) (string, error) {
	if v4 {
		return uuid.New().String(), nil
	}
	return uuid.NewV7().String(), nil
}

func GenerateShortUUIDString(v4 bool) (string, error) {
	if v4 {
		return GenerateRUIDString(18)
	}
	str := uuid.NewV7().String()
	return str[15:18] + str[20:23] + str[24:], nil
}

func GenerateRUIDString(length int) (string, error) {
	if length <= 0 || length > 30 {
		length = 18
	}
	str := uuid.New().String()
	shortUUID := str[0:8] + str[9:13] + str[15:18] + str[20:23] + str[24:]
	return shortUUID[:length], nil
}
