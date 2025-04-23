package anbuCrypto

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	u "github.com/tanq16/anbu/utils"
)

func JwtParse(tokenString string) {
	fmt.Println()
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		u.PrintError("invalid token format")
	}
	header, err := jwtDecodeSegment(parts[0])
	if err != nil {
		u.PrintError(fmt.Sprintf("failed to decode header: %v", err))
	}
	payload, err := jwtDecodeSegment(parts[1])
	if err != nil {
		u.PrintError(fmt.Sprintf("failed to decode payload: %v", err))
	}

	// Print the header and payload in a table format
	headerTable := u.NewTable([]string{"Header", "Value"})
	for k, v := range header {
		headerTable.Rows = append(headerTable.Rows, []string{k, fmt.Sprintf("%v", v)})
	}
	payloadTable := u.NewTable([]string{"Payload", "Value"})
	for k, v := range payload {
		payloadTable.Rows = append(payloadTable.Rows, []string{k, fmt.Sprintf("%v", v)})
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
