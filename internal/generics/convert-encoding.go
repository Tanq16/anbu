package anbuGenerics

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
)

func textToBase64(input string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	fmt.Println(encoded)
}

func base64ToText(input string) {
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode base64")
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
		log.Fatal().Err(err).Msg("Failed to decode hex")
		return
	}
	fmt.Println(string(decoded))
}

func base64ToHex(input string) {
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode base64")
		return
	}
	hexEncoded := hex.EncodeToString(decoded)
	fmt.Println(hexEncoded)
}

func hexToBase64(input string) {
	decoded, err := hex.DecodeString(strings.TrimSpace(input))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode hex")
		return
	}
	base64Encoded := base64.StdEncoding.EncodeToString(decoded)
	fmt.Println(base64Encoded)
}

func urlToText(input string) {
	decoded, err := url.QueryUnescape(input)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode URL")
		return
	}
	fmt.Println(decoded)
}

func textToUrl(input string) {
	encoded := url.QueryEscape(input)
	fmt.Println(encoded)
}

func jwtDecode(tokenString string) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		log.Error().Msg("invalid token format")
	}
	header, err := jwtDecodeSegment(parts[0])
	if err != nil {
		log.Error().Err(err).Msg("failed to decode header")
	}
	payload, err := jwtDecodeSegment(parts[1])
	if err != nil {
		log.Error().Err(err).Msg("failed to decode payload")
	}

	// Print the header and payload in a table format
	headerTable := u.NewTable([]string{"Header", "Value"})
	for k, v := range header {
		switch v := v.(type) {
		case float64:
			headerTable.Rows = append(headerTable.Rows, []string{k, fmt.Sprintf("%.0f", v)})
		case int64:
			headerTable.Rows = append(headerTable.Rows, []string{k, fmt.Sprintf("%d", v)})
		default:
			headerTable.Rows = append(headerTable.Rows, []string{k, fmt.Sprintf("%v", v)})
		}
	}
	payloadTable := u.NewTable([]string{"Payload", "Value"})
	for k, v := range payload {
		switch v := v.(type) {
		case float64:
			payloadTable.Rows = append(payloadTable.Rows, []string{k, fmt.Sprintf("%.0f", v)})
		case int64:
			payloadTable.Rows = append(payloadTable.Rows, []string{k, fmt.Sprintf("%d", v)})
		default:
			payloadTable.Rows = append(payloadTable.Rows, []string{k, fmt.Sprintf("%v", v)})
		}
	}
	headerTable.PrintTable(false)
	payloadTable.PrintTable(false)
}

func jwtDecodeSegment(seg string) (u.Dictionary, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}
	bytes, err := base64.URLEncoding.DecodeString(seg)
	if err != nil {
		return nil, err
	}
	var result u.Dictionary
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}
