package tools

import (
	"fmt"
	"strings"
	"time"
)

type TimeFormats struct {
	Epoch        int64  `json:"epoch"`
	EpochMS      int64  `json:"epoch_ms"`
	RFC822Local  string `json:"rfc822_local"`
	ISO8601Local string `json:"iso8601_local"`
	ISO8601UTC   string `json:"iso8601_utc"`
	HumanUTC     string `json:"human_utc"`
}

type TimeUntilResult struct {
	Target  string `json:"target"`
	Now     string `json:"now"`
	Seconds int64  `json:"seconds"`
	Future  bool   `json:"future"`
	Human   string `json:"human"`
}

type EpochDiffResult struct {
	Seconds int64   `json:"seconds"`
	Minutes float64 `json:"minutes"`
	Hours   float64 `json:"hours"`
	Days    float64 `json:"days"`
	Human   string  `json:"human"`
}

const humanLayout = "Mon Jan 2 15:04:05 MST 2006"

func FormatTime(t time.Time) TimeFormats {
	return TimeFormats{
		Epoch:        t.Unix(),
		EpochMS:      t.UnixMilli(),
		RFC822Local:  t.Local().Format(time.RFC822),
		ISO8601Local: t.Local().Format(time.RFC3339),
		ISO8601UTC:   t.UTC().Format(time.RFC3339),
		HumanUTC:     t.UTC().Format(humanLayout),
	}
}

func Until(target, now time.Time) TimeUntilResult {
	diff := target.Sub(now)
	future := diff > 0
	if !future {
		diff = -diff
	}
	return TimeUntilResult{
		Target:  target.Format(time.RFC3339),
		Now:     now.Format(time.RFC3339),
		Seconds: int64(diff.Seconds()),
		Future:  future,
		Human:   FormatDuration(diff),
	}
}

func EpochDiff(d time.Duration) EpochDiffResult {
	abs := d
	if abs < 0 {
		abs = -abs
	}
	return EpochDiffResult{
		Seconds: int64(d.Seconds()),
		Minutes: d.Minutes(),
		Hours:   d.Hours(),
		Days:    d.Hours() / 24,
		Human:   FormatDuration(abs),
	}
}

func FormatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	var parts []string
	for _, unit := range []struct {
		n    int
		name string
	}{{days, "day"}, {hours, "hour"}, {minutes, "minute"}} {
		if unit.n > 0 {
			parts = append(parts, plural(unit.n, unit.name))
		}
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, plural(seconds, "second"))
	}
	return strings.Join(parts, ", ")
}

func plural(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}
