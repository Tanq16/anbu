package anbuNetwork

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tanq16/anbu/utils"
)

func GetPublicIP() (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("https://ipinfo.io")
	if err != nil {
		return "", fmt.Errorf("failed to connect to ipinfo.io: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	var data utils.Dictionary
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}
	ip := data.UnwindString("ip")
	if ip == "" {
		return "", fmt.Errorf("no IP address found in the response")
	}
	return ip, nil
}

// func GetLocalIP() (string, error) {
// 	// gets the local IP address, hostname, DNS servers, and subnet mask
// 	return ip, nil
// }
