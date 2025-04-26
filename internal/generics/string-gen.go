package anbuGenerics

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	u "github.com/tanq16/anbu/utils"
)

// generates a random string of specified length
func GenerateRandomString(length int) {
	if length <= 0 {
		length = 100
	}
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		u.PrintError(fmt.Sprintf("failed to generate random bytes: %v", err))
		return
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

// generates a sequence of alphabetic characters
func GenerateSequenceString(length int) {
	if length <= 0 {
		u.PrintWarning("length must be greater than 0; using 100")
		length = 100
	}
	alphabet := "abcdefghijklmnopqrstuvxyz"
	var result strings.Builder
	for result.Len() < length {
		result.WriteString(alphabet)
	}
	fmt.Println(result.String()[:length])
}

// repeats a string a specified number of times
func GenerateRepetitionString(count int, str string) {
	if count <= 0 {
		u.PrintWarning("count must be greater than 0; using 10")
		count = 10
	}
	var result strings.Builder
	for range count {
		result.WriteString(str)
	}
	fmt.Println(result.String())
}

// use google/uuid package to generate a UUID
func GenerateUUIDString() {
	uuid, _ := uuid.NewRandom()
	fmt.Println(uuid.String())
}

// generates shorter UUID string
func GenerateRUIDString(len string) {
	length, err := strconv.Atoi(len)
	if err != nil {
		u.PrintError("not a valid length")
		return
	}
	if length <= 0 || length > 30 {
		u.PrintWarning("length must be between 1 and 30; using 18")
		length = 18
	}
	uuid, _ := uuid.NewRandom()
	// remove version and variant bits from UUID
	shortUUID := uuid.String()[0:8] + uuid.String()[9:13] + uuid.String()[15:18] + uuid.String()[20:23] + uuid.String()[24:]
	fmt.Println(shortUUID[:length])
}
