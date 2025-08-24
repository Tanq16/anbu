package anbuGenerics

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func GenerateRandomString(length int) {
	if length <= 0 {
		length = 100
	}
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to generate random bytes")
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
	fmt.Println(encoded)
}

func GenerateSequenceString(length int) {
	if length <= 0 {
		log.Warn().Msg("length must be greater than 0; using 100")
		length = 100
	}
	alphabet := "abcdefghijklmnopqrstuvxyz"
	var result strings.Builder
	for result.Len() < length {
		result.WriteString(alphabet)
	}
	fmt.Println(result.String()[:length])
}

func GenerateRepetitionString(count int, str string) {
	if count <= 0 {
		log.Warn().Msg("count must be greater than 0; using 10")
		count = 10
	}
	var result strings.Builder
	for range count {
		result.WriteString(str)
	}
	fmt.Println(result.String())
}

func GenerateUUIDString() {
	uuid, _ := uuid.NewRandom()
	fmt.Println(uuid.String())
}

// generates shorter UUID string
func GenerateRUIDString(len string) {
	length, err := strconv.Atoi(len)
	if err != nil {
		log.Fatal().Msg("not a valid length")
	}
	if length <= 0 || length > 30 {
		log.Warn().Msg("length must be between 1 and 30; using 18")
		length = 18
	}
	uuid, _ := uuid.NewRandom()
	// remove version and variant bits from UUID
	shortUUID := uuid.String()[0:8] + uuid.String()[9:13] + uuid.String()[15:18] + uuid.String()[20:23] + uuid.String()[24:]
	fmt.Println(shortUUID[:length])
}
