package anbuGenerics

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	u "github.com/tanq16/anbu/utils"
)

func printTimeTable(concern time.Time) {
	utcTime := concern.UTC()
	localTime := concern.Local()
	table := u.NewTable([]string{"Format", "Value"})
	table.Rows = [][]string{
		{"Epoch", strconv.FormatInt(concern.Unix(), 10)},
		{"RFC 822 human local", localTime.Format(time.RFC822)},
		{"ISO 8601 local", localTime.Format(time.RFC3339)},
		{"ISO 8601 UTC", utcTime.Format(time.RFC3339)},
		{"Human UTC", utcTime.Format("Mon Jan 2 15:04:05 MST 2006")},
	}
	table.PrintTable()
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
	u.LineBreak()
	u.PrintGeneric(fmt.Sprintf("Target time: %s", u.FDebug(targetTime.Format("Mon Jan 2 15:04:05 MST 2006"))))
	u.PrintGeneric(fmt.Sprintf("Current time: %s", u.FDebug(now.Format("Mon Jan 2 15:04:05 MST 2006"))))
	u.LineBreak()
	if direction == "until" {
		u.PrintGeneric(fmt.Sprintf("Target time is %s from now", u.FInfo(timeFormatDuration(diff))))
	} else {
		u.PrintGeneric(fmt.Sprintf("Target time was %s ago", u.FInfo(timeFormatDuration(diff))))
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

func parseTime(timeStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC822,
		time.RFC1123,
		time.UnixDate,
		time.DateTime,
		"Mon Jan 2 15:04:05 MST 2006",
		"January 2, 2006 3:04:05 PM MST",
		"2006-01-02",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"02-Jan-2006",
		"2006-01-02T15:04:05Z07:00",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}
	epoch, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse time string with any known format: %w", err)
	}
	return time.Unix(epoch, 0), nil
}

func TimeParse(timeStr string) error {
	parsedTime, err := parseTime(timeStr)
	if err != nil {
		return err
	}
	printTimeTable(parsedTime)
	return nil
}

func TimeUntil(timeStr string) error {
	parsedTime, err := parseTime(timeStr)
	if err != nil {
		return err
	}
	printTimeDifferenceFromNow(parsedTime)
	return nil
}

func TimeCurrent() {
	printTimeTable(time.Now())
}

func TimeEpochDiff(epochs []int64) {
	var epoch1, epoch2 int64
	if len(epochs) == 1 {
		epoch1, epoch2 = epochs[0], time.Now().Unix()
	} else {
		epoch1, epoch2 = epochs[0], epochs[1]
	}
	t1 := time.Unix(epoch1, 0)
	t2 := time.Unix(epoch2, 0)
	diff := t2.Sub(t1)
	u.PrintGeneric("Time difference:")
	u.PrintGeneric(fmt.Sprintf("  %s  %d", u.FSuccess("Seconds:"), int64(diff.Seconds())))
	u.PrintGeneric(fmt.Sprintf("  %s  %.1f", u.FSuccess("Minutes:"), diff.Minutes()))
	u.PrintGeneric(fmt.Sprintf("  %s  %.1f", u.FSuccess("Hours:"), diff.Hours()))
	u.PrintGeneric(fmt.Sprintf("  %s  %.1f", u.FSuccess("Days:"), diff.Hours()/24))
	if diff > 0 {
		u.PrintGeneric(fmt.Sprintf("\n%s is %s after %s", u.FInfo("Time 2"), u.FSuccess(timeFormatDuration(diff)), u.FInfo("Time 1")))
	} else {
		u.PrintGeneric(fmt.Sprintf("\n%s is %s before %s", u.FInfo("Time 2"), u.FSuccess(timeFormatDuration(-diff)), u.FInfo("Time 1")))
	}
}
