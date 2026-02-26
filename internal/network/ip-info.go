package anbuNetwork

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/internal/utils"
)

type NetworkInterface struct {
	Name       string
	IPv4Addr   string
	IPv4Mask   string
	IPv6Addr   string
	MACAddr    string
	IsUp       bool
	MTU        int
	IsLoopback bool
}

func GetLocalIPInfo(includeIPv6 bool) {
	interfaces, err := net.Interfaces()
	if err != nil {
		u.PrintError("failed to get network interfaces", err)
		return
	}
	ipv4Table := u.NewTable([]string{"Interface", "IP Address", "Subnet Mask", "MAC Address", "Status"})
	ipv6Table := u.NewTable([]string{"Interface", "IPv6 Address", "MAC Address", "Status"})

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
			ipv4Table.Rows = append(ipv4Table.Rows, []string{
				iface.Name, ipv4, mask, iface.HardwareAddr.String(), status,
			})
		}
		if ipv6 != "" {
			ipv6Table.Rows = append(ipv6Table.Rows, []string{
				iface.Name, ipv6, iface.HardwareAddr.String(), status,
			})
		}
	}

	u.LineBreak()
	ipv4Table.PrintTable(false)
	if includeIPv6 {
		ipv6Table.PrintTable(false)
	}
	u.LineBreak()

	publicIP, err := GetPublicIP()
	pubIPTable := u.NewTable([]string{"Field", "Value"})
	if err == nil && publicIP != nil {
		geography := struct {
			Country  string
			Region   string
			City     string
			Postal   string
			Timezone string
		}{}
		for key, value := range publicIP {
			if key == "readme" {
				continue
			}
			if key == "loc" {
				continue
			}
			if key == "country" {
				if s, ok := value.(string); ok {
					geography.Country = s
				}
				continue
			}
			if key == "region" {
				if s, ok := value.(string); ok {
					geography.Region = s
				}
				continue
			}
			if key == "city" {
				if s, ok := value.(string); ok {
					geography.City = s
				}
				continue
			}
			if key == "postal" {
				if s, ok := value.(string); ok {
					geography.Postal = s
				}
				continue
			}
			if key == "timezone" {
				if s, ok := value.(string); ok {
					geography.Timezone = s
				}
				continue
			}
			pubIPTable.Rows = append(pubIPTable.Rows, []string{key, fmt.Sprintf("%v", value)})
		}
		pubIPTable.Rows = append(pubIPTable.Rows, []string{"geography", fmt.Sprintf("%s, %s, %s, %s (TZ: %s)", geography.Postal, geography.City, geography.Region, geography.Country, geography.Timezone)})
		pubIPTable.PrintTable(false)
	} else {
		u.PrintWarn("Could not retrieve public IP", err)
	}
}

func GetPublicIP() (u.Dictionary, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("https://ipinfo.io")
	log.Debug().Str("package", "network").Msg("requested IP info from IP-Info.io")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ipinfo.io: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var data u.Dictionary
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}
	return data, nil
}
