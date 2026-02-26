package anbuGenerics

import (
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/google/uuid"
	u "github.com/tanq16/anbu/internal/utils"
)

func GenerateRandomString(length int) {
	if length <= 0 {
		length = 100
	}
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		u.PrintFatal("failed to generate random bytes", err)
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
	u.PrintGeneric(encoded)
}

func GenerateSequenceString(length int) {
	if length <= 0 {
		u.PrintWarn("length must be greater than 0; using 100", nil)
		length = 100
	}
	alphabet := "abcdefghijklmnopqrstuvxyz"
	var result strings.Builder
	for result.Len() < length {
		result.WriteString(alphabet)
	}
	u.PrintGeneric(result.String()[:length])
}

func GenerateRepetitionString(count int, str string) {
	if count <= 0 {
		u.PrintWarn("count must be greater than 0; using 10", nil)
		count = 10
	}
	var result strings.Builder
	for range count {
		result.WriteString(str)
	}
	u.PrintGeneric(result.String())
}

func GenerateUUIDString() {
	uuid, _ := uuid.NewRandom()
	u.PrintGeneric(uuid.String())
}

func GenerateRUIDString(len string) {
	length, err := strconv.Atoi(len)
	if err != nil {
		u.PrintFatal("not a valid length", err)
	}
	if length <= 0 || length > 30 {
		u.PrintWarn("length must be between 1 and 30; using 18", nil)
		length = 18
	}
	uuid, _ := uuid.NewRandom()
	shortUUID := uuid.String()[0:8] + uuid.String()[9:13] + uuid.String()[15:18] + uuid.String()[20:23] + uuid.String()[24:]
	u.PrintGeneric(shortUUID[:length])
}
