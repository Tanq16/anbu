package genericsCmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
)

var TimeCmd = &cobra.Command{
	Use:     "time",
	Aliases: []string{"t"},
	Short:   "Show times in common formats, parse a string, or diff epochs",
}

var timeNowCmd = &cobra.Command{
	Use:   "now",
	Short: "Print the current time in common formats",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		printTimeTable(time.Now())
	},
}

var timeParseCmd = &cobra.Command{
	Use:   "parse <time-str>",
	Short: "Parse a time string and print it in common formats",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		parsed, err := anbuGenerics.ParseTime(args[0])
		if err != nil {
			u.PrintFatal("could not parse time", err)
		}
		printTimeTable(parsed)
	},
}

var timeUntilCmd = &cobra.Command{
	Use:   "until <time-str>",
	Short: "Print how far a time is from now",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target, now, err := anbuGenerics.TimeUntil(args[0])
		if err != nil {
			u.PrintFatal("could not parse time", err)
		}
		printTimeUntil(target, now)
	},
}

var timeDiffCmd = &cobra.Command{
	Use:   "diff <epoch> [epoch]",
	Short: "Print the difference between Unix epochs",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		epochs := make([]int64, len(args))
		for i, arg := range args {
			epoch, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				u.PrintFatal("invalid epoch", err)
			}
			epochs[i] = epoch
		}
		printEpochDiff(anbuGenerics.TimeEpochDiff(epochs))
	},
}

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

func printTimeUntil(targetTime, now time.Time) {
	var diff time.Duration
	future := targetTime.After(now)
	if future {
		diff = targetTime.Sub(now)
	} else {
		diff = now.Sub(targetTime)
	}
	u.LineBreak()
	u.PrintGeneric(fmt.Sprintf("Target time: %s", u.FDebug(targetTime.Format("Mon Jan 2 15:04:05 MST 2006"))))
	u.PrintGeneric(fmt.Sprintf("Current time: %s", u.FDebug(now.Format("Mon Jan 2 15:04:05 MST 2006"))))
	u.LineBreak()
	if future {
		u.PrintGeneric(fmt.Sprintf("Target time is %s from now", u.FInfo(formatDuration(diff))))
		return
	}
	u.PrintGeneric(fmt.Sprintf("Target time was %s ago", u.FInfo(formatDuration(diff))))
}

func printEpochDiff(diff time.Duration) {
	u.PrintGeneric("Time difference:")
	u.PrintGeneric(fmt.Sprintf("  %s  %d", u.FSuccess("Seconds:"), int64(diff.Seconds())))
	u.PrintGeneric(fmt.Sprintf("  %s  %.1f", u.FSuccess("Minutes:"), diff.Minutes()))
	u.PrintGeneric(fmt.Sprintf("  %s  %.1f", u.FSuccess("Hours:"), diff.Hours()))
	u.PrintGeneric(fmt.Sprintf("  %s  %.1f", u.FSuccess("Days:"), diff.Hours()/24))
	if diff > 0 {
		u.PrintGeneric(fmt.Sprintf("\n%s is %s after %s", u.FInfo("Time 2"), u.FSuccess(formatDuration(diff)), u.FInfo("Time 1")))
		return
	}
	u.PrintGeneric(fmt.Sprintf("\n%s is %s before %s", u.FInfo("Time 2"), u.FSuccess(formatDuration(-diff)), u.FInfo("Time 1")))
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	var parts []string
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

func init() {
	TimeCmd.AddCommand(timeNowCmd)
	TimeCmd.AddCommand(timeParseCmd)
	TimeCmd.AddCommand(timeUntilCmd)
	TimeCmd.AddCommand(timeDiffCmd)
}
