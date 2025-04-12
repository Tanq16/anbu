package anbuTime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tanq16/anbu/utils"
)

type timeFormat struct {
	Format string
	Value  string
}

func printTimeTable(concern time.Time) error {
	utcTime := concern.UTC()
	localTime := concern.Local()
	// Create table for parsed time formats
	table := utils.MarkdownTable{
		Headers: []string{"Format", "Value"},
		Rows:    [][]string{},
	}
	timeFormats := []timeFormat{
		{"ISO8601 UTC", utcTime.Format(time.RFC3339)},
		{"Human UTC", utcTime.Format("Mon Jan 2 15:04:05 MST 2006")},
		{"ISO8601 Local", localTime.Format(time.RFC3339)},
		{"Human Local", localTime.Format("Mon Jan 2 15:04:05 MST 2006")},
		{"RFC822", localTime.Format(time.RFC822)},
		{"Epoch", fmt.Sprintf("%d", concern.Unix())},
		{"Epoch Nano", fmt.Sprintf("%d", concern.UnixNano())},
		{"Date Only", localTime.Format("2006-01-02")},
		{"Time Only", localTime.Format("15:04:05")},
		{"Database", localTime.Format("2006-01-02 15:04:05")},
	}
	for _, format := range timeFormats {
		table.Rows = append(table.Rows, []string{format.Format, format.Value})
	}
	return table.OutMDPrint(false)
}

func printTimeTablePurple(concern time.Time) error {
	utcTime := concern.UTC()
	table := utils.MarkdownTable{
		Headers: []string{"Item", "Value"},
		Rows:    [][]string{},
	}
	formats := []timeFormat{
		{"ISO8601 UTC", utcTime.Format(time.RFC3339)},
		{"Human UTC", utcTime.Format("Mon Jan 2 15:04:05 MST 2006")},
		{"RFC822", concern.Format(time.RFC822)},
		{"Epoch", fmt.Sprintf("%d", concern.Unix())},
	}
	for _, format := range formats {
		table.Rows = append(table.Rows, []string{format.Format, format.Value})
	}
	ipAddr, err := getPublicIP()
	if err != nil {
		log.Warn().Err(err).Msg("could not get public IP address")
	} else {
		table.Rows = append(table.Rows, []string{"Public IP", ipAddr})
	}
	table.OutMDPrint(false)
	return nil
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	parts := []string{}
	if days > 0 {
		if days == 1 {
			parts = append(parts, "1 day")
		} else {
			parts = append(parts, fmt.Sprintf("%d days", days))
		}
	}
	if hours > 0 {
		if hours == 1 {
			parts = append(parts, "1 hour")
		} else {
			parts = append(parts, fmt.Sprintf("%d hours", hours))
		}
	}
	if minutes > 0 {
		if minutes == 1 {
			parts = append(parts, "1 minute")
		} else {
			parts = append(parts, fmt.Sprintf("%d minutes", minutes))
		}
	}
	return strings.Join(parts, ", ")
}

func getPublicIP() (string, error) {
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

func Parse(timeStr string, printType string) error {
	formats := []string{
		time.RFC3339,
		time.RFC822,
		time.RFC1123,
		time.UnixDate,
		time.DateTime,
		"Mon Jan 2 15:04:05 MST 2006",    // Human readable format
		"January 2, 2006 3:04:05 PM MST", // should read "March 8, 2025 14:05:43 GMT-4"
		"2006-01-02",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"02-Jan-2006",
		"2006-01-02T15:04:05Z07:00", // Additional ISO8601 variant
	}
	var parsedTime time.Time
	// Try to parse with each common format
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			parsedTime = t.UTC()
			break
		}
	}
	if parsedTime.IsZero() {
		// check for epoch format
		checkEpock, err := strconv.ParseInt(timeStr, 10, 64)
		if err == nil {
			parsedTime = time.Unix(checkEpock, 0)
		} else {
			return fmt.Errorf("could not parse time string with any known format")
		}
	}
	var err error
	if printType == "normal" {
		err = printTimeTable(parsedTime)
	} else if printType == "purple" {
		err = printTimeTablePurple(parsedTime)
	}
	if err != nil {
		return fmt.Errorf("could not print time table: %w", err)
	}
	return nil
}

func Current() {
	currentTime := time.Now()
	if err := printTimeTable(currentTime); err != nil {
		log.Error().Err(err).Msg("could not print table")
	}
}

func Purple() {
	currentTime := time.Now()
	if err := printTimeTablePurple(currentTime); err != nil {
		log.Error().Err(err).Msg("could not print table")
	}
}

func EpochDiff(epochs []int64) error {
	var epoch1, epoch2 int64
	if len(epochs) == 0 {
		return fmt.Errorf("no epochs provided")
	} else if len(epochs) == 1 {
		epoch1, epoch2 = epochs[0], time.Now().Unix()
	} else {
		epoch1, epoch2 = epochs[0], epochs[1]
	}
	// Convert to time.Time for better manipulation
	t1 := time.Unix(epoch1, 0)
	t2 := time.Unix(epoch2, 0)
	diff := t2.Sub(t1)
	// Show difference in multiple units
	fmt.Printf("Time difference:\n")
	fmt.Printf("  Seconds:  %d\n", int64(diff.Seconds()))
	fmt.Printf("  Minutes:  %.1f\n", diff.Minutes())
	fmt.Printf("  Hours:    %.1f\n", diff.Hours())
	fmt.Printf("  Days:     %.1f\n", diff.Hours()/24)
	// Add human readable description
	if diff > 0 {
		fmt.Printf("\nTime 2 is %s after Time 1\n", formatDuration(diff))
	} else {
		fmt.Printf("\nTime 2 is %s before Time 1\n", formatDuration(-diff))
	}
	return nil
}
