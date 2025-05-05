package anbuGenerics

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	u "github.com/tanq16/anbu/utils"
)

func textToBase64(input string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	fmt.Println(encoded)
}

func base64ToText(input string) {
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		u.PrintError(fmt.Sprintf("Failed to decode base64: %v", err))
		return
	}
	fmt.Println(string(decoded))
}

func textToHex(input string) {
	encoded := hex.EncodeToString([]byte(input))
	fmt.Println(encoded)
}

func hexToText(input string) {
	decoded, err := hex.DecodeString(strings.TrimSpace(input))
	if err != nil {
		u.PrintError(fmt.Sprintf("Failed to decode hex: %v", err))
		return
	}
	fmt.Println(string(decoded))
}

func base64ToHex(input string) {
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		u.PrintError(fmt.Sprintf("Failed to decode base64: %v", err))
		return
	}
	hexEncoded := hex.EncodeToString(decoded)
	fmt.Println(hexEncoded)
}

func hexToBase64(input string) {
	decoded, err := hex.DecodeString(strings.TrimSpace(input))
	if err != nil {
		u.PrintError(fmt.Sprintf("Failed to decode hex: %v", err))
		return
	}
	base64Encoded := base64.StdEncoding.EncodeToString(decoded)
	fmt.Println(base64Encoded)
}

func urlToText(input string) {
	decoded, err := url.QueryUnescape(input)
	if err != nil {
		u.PrintError(fmt.Sprintf("Failed to decode URL: %v", err))
		return
	}
	fmt.Println(decoded)
}

func textToUrl(input string) {
	encoded := url.QueryEscape(input)
	fmt.Println(encoded)
}
