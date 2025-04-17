package anbuGenerics

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tanq16/anbu/utils"
)

func JwtParse(tokenString string) error {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid token format")
	}
	header, err := jwtDecodeSegment(parts[0])
	if err != nil {
		return fmt.Errorf("failed to decode header: %v", err)
	}
	payload, err := jwtDecodeSegment(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode payload: %v", err)
	}
	log.Debug().Interface("header", header).Interface("payload", payload).Msg("decoded token")
	// Print the header and payload in a table format
	headerTable := utils.MarkdownTable{
		Headers: []string{"Header", "Value"},
		Rows:    [][]string{},
	}
	for k, v := range header {
		headerTable.Rows = append(headerTable.Rows, []string{k, fmt.Sprintf("%v", v)})
	}
	payloadTable := utils.MarkdownTable{
		Headers: []string{"Payload", "Value"},
		Rows:    [][]string{},
	}
	for k, v := range payload {
		payloadTable.Rows = append(payloadTable.Rows, []string{k, fmt.Sprintf("%v", v)})
	}
	if err := headerTable.OutMDPrint(false); err != nil {
		return fmt.Errorf("failed to print header table: %v", err)
	}
	if err := payloadTable.OutMDPrint(false); err != nil {
		return fmt.Errorf("failed to print payload table: %v", err)
	}
	return nil
}

func jwtDecodeSegment(seg string) (utils.Dictionary, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}
	bytes, err := base64.URLEncoding.DecodeString(seg)
	if err != nil {
		return nil, err
	}
	var result utils.Dictionary
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}
