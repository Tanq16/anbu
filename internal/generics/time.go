package anbuGenerics

import (
	"fmt"
	"strconv"
	"time"
)

func ParseTime(timeStr string) (time.Time, error) {
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

func TimeUntil(timeStr string) (target time.Time, now time.Time, err error) {
	target, err = ParseTime(timeStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return target, time.Now(), nil
}

func TimeEpochDiff(epochs []int64) time.Duration {
	var epoch1, epoch2 int64
	if len(epochs) == 1 {
		epoch1, epoch2 = epochs[0], time.Now().Unix()
	} else {
		epoch1, epoch2 = epochs[0], epochs[1]
	}
	return time.Unix(epoch2, 0).Sub(time.Unix(epoch1, 0))
}
