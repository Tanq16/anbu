package anbuGenerics

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	u "github.com/tanq16/anbu/utils"
)

func urlToText(input string) error {
	decoded, err := url.QueryUnescape(input)
	if err != nil {
		return fmt.Errorf("failed to decode URL: %w", err)
	}
	u.PrintGeneric(decoded)
	return nil
}

func textToUrl(input string) error {
	encoded := url.QueryEscape(input)
	u.PrintGeneric(encoded)
	return nil
}

func jwtDecode(tokenString string) error {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid token format: expected 3 parts separated by '.'")
	}
	header, err := jwtDecodeSegment(parts[0])
	if err != nil {
		return fmt.Errorf("failed to decode header: %w", err)
	}
	payload, err := jwtDecodeSegment(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode payload: %w", err)
	}

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
	return nil
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
