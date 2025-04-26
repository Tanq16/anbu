package anbuGenerics

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	anbuNetwork "github.com/tanq16/anbu/internal/network"
	u "github.com/tanq16/anbu/utils"
)

type timeFormat struct {
	Format string
	Value  string
}

func printTimeTable(concern time.Time) {
	utcTime := concern.UTC()
	localTime := concern.Local()
	// Create table for parsed time formats
	table := u.NewTable([]string{"Format", "Value"})
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
	table.PrintTable(false)
}

func printTimeTablePurple(concern time.Time) {
	utcTime := concern.UTC()
	table := u.NewTable([]string{"Item", "Value"})
	formats := []timeFormat{
		{"ISO8601 UTC", utcTime.Format(time.RFC3339)},
		{"Human UTC", utcTime.Format("Mon Jan 2 15:04:05 MST 2006")},
		{"RFC822", concern.Format(time.RFC822)},
		{"Epoch", fmt.Sprintf("%d", concern.Unix())},
	}
	for _, format := range formats {
		table.Rows = append(table.Rows, []string{format.Format, format.Value})
	}
	ipAddr, err := anbuNetwork.GetPublicIP()
	if err != nil {
		u.PrintWarning("could not get public IP address")
	} else {
		ipAddress := ipAddr.UnwindString("ip")
		table.Rows = append(table.Rows, []string{"Public IP", ipAddress})
	}
	table.PrintTable(false)
}

func printTimeDifferenceFromNow(targetTime time.Time) {
	now := time.Now()
	var diff time.Duration
	var direction string
	if targetTime.After(now) {
		diff = targetTime.Sub(now)
		direction = "until"
	} else {
		diff = now.Sub(targetTime)
		direction = "ago"
	}
	fmt.Println()
	fmt.Printf("%s: %s\n", u.FDetail("Target time"), u.FDebug(targetTime.Format("Mon Jan 2 15:04:05 MST 2006")))
	fmt.Printf("%s: %s\n", u.FDetail("Current time"), u.FDebug(now.Format("Mon Jan 2 15:04:05 MST 2006")))
	fmt.Println()
	// Print direction-aware message
	if direction == "until" {
		fmt.Printf("Target time is %s from now\n", u.FDetail(timeFormatDuration(diff)))
	} else {
		fmt.Printf("Target time was %s ago\n", u.FDetail(timeFormatDuration(diff)))
	}
}

func timeFormatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
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
	if seconds > 0 || len(parts) == 0 {
		if seconds == 1 {
			parts = append(parts, "1 second")
		} else {
			parts = append(parts, fmt.Sprintf("%d seconds", seconds))
		}
	}
	return strings.Join(parts, ", ")
}

func TimeParse(timeStr string, printType string) {
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
			parsedTime = t
			break
		}
	}
	if parsedTime.IsZero() {
		// check for epoch format
		checkEpock, err := strconv.ParseInt(timeStr, 10, 64)
		if err == nil {
			parsedTime = time.Unix(checkEpock, 0)
		} else {
			u.PrintError("could not parse time string with any known format")
			return
		}
	}
	switch printType {
	case "normal":
		printTimeTable(parsedTime)
	case "purple":
		printTimeTablePurple(parsedTime)
	case "diff":
		printTimeDifferenceFromNow(parsedTime)
	default:
		printTimeTable(parsedTime)
	}
}

func TimeCurrent() {
	currentTime := time.Now()
	printTimeTable(currentTime)
}

func TimePurple() {
	currentTime := time.Now()
	printTimeTablePurple(currentTime)
}

func TimeEpochDiff(epochs []int64) {
	var epoch1, epoch2 int64
	if len(epochs) == 0 {
		u.PrintError("No epochs provided")
		return
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
	fmt.Println(u.FDetail("Time difference:"))
	fmt.Printf("  %s  %d\n", u.FSuccess("Seconds:"), int64(diff.Seconds()))
	fmt.Printf("  %s  %.1f\n", u.FSuccess("Minutes:"), diff.Minutes())
	fmt.Printf("  %s  %.1f\n", u.FSuccess("Hours:"), diff.Hours())
	fmt.Printf("  %s  %.1f\n", u.FSuccess("Days:"), diff.Hours()/24)
	// Add human readable description
	if diff > 0 {
		fmt.Printf("\n%s is %s after %s\n", u.FInfo("Time 2"), u.FSuccess(timeFormatDuration(diff)), u.FInfo("Time 1"))
	} else {
		fmt.Printf("\n%s is %s before %s\n", u.FInfo("Time 2"), u.FSuccess(timeFormatDuration(-diff)), u.FInfo("Time 1"))
	}
}
