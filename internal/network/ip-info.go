package anbuNetwork

import (
	"encoding/json/v2"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type IPInfo struct {
	IPv4      [][]string
	IPv6      [][]string
	Public    [][]string
	PublicErr error
}

func GetLocalIPInfo() (IPInfo, error) {
	var info IPInfo
	interfaces, err := net.Interfaces()
	if err != nil {
		return info, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}
		var ipv4, mask, ipv6 string
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if ip4 := v.IP.To4(); ip4 != nil {
					ipv4 = ip4.String()
					mask = net.IP(v.Mask).String()
				} else if ip6 := v.IP.To16(); ip6 != nil && !ip6.Equal(v.IP.To4()) {
					ipv6 = ip6.String()
				}
			}
		}
		status := "Up"
		if iface.Flags&net.FlagLoopback != 0 {
			status = "Loopback"
		}
		if ipv4 != "" {
			info.IPv4 = append(info.IPv4, []string{
				iface.Name, ipv4, mask, iface.HardwareAddr.String(), status,
			})
		}
		if ipv6 != "" {
			info.IPv6 = append(info.IPv6, []string{
				iface.Name, ipv6, iface.HardwareAddr.String(), status,
			})
		}
	}

	publicIP, err := getPublicIP()
	if err != nil {
		info.PublicErr = err
		return info, nil
	}
	info.Public = publicRows(publicIP)
	return info, nil
}

func publicRows(publicIP map[string]any) [][]string {
	var rows [][]string
	geography := struct {
		Country  string
		Region   string
		City     string
		Postal   string
		Timezone string
	}{}
	for key, value := range publicIP {
		switch key {
		case "readme", "loc":
			continue
		case "country":
			if s, ok := value.(string); ok {
				geography.Country = s
			}
		case "region":
			if s, ok := value.(string); ok {
				geography.Region = s
			}
		case "city":
			if s, ok := value.(string); ok {
				geography.City = s
			}
		case "postal":
			if s, ok := value.(string); ok {
				geography.Postal = s
			}
		case "timezone":
			if s, ok := value.(string); ok {
				geography.Timezone = s
			}
		default:
			rows = append(rows, []string{key, fmt.Sprintf("%v", value)})
		}
	}
	rows = append(rows, []string{"geography", fmt.Sprintf("%s, %s, %s, %s (TZ: %s)", geography.Postal, geography.City, geography.Region, geography.Country, geography.Timezone)})
	return rows
}

func getPublicIP() (map[string]any, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("https://ipinfo.io")
	log.Debug().Msg("requested IP info from IP-Info.io")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ipinfo.io: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}
	return data, nil
}
