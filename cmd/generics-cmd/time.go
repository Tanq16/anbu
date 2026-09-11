package genericsCmd

import (
	"strconv"

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
		anbuGenerics.TimeCurrent()
	},
}

var timeParseCmd = &cobra.Command{
	Use:   "parse <time-str>",
	Short: "Parse a time string and print it in common formats",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.TimeParse(args[0]); err != nil {
			u.PrintFatal("could not parse time", err)
		}
	},
}

var timeUntilCmd = &cobra.Command{
	Use:   "until <time-str>",
	Short: "Print how far a time is from now",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.TimeUntil(args[0]); err != nil {
			u.PrintFatal("could not parse time", err)
		}
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
		anbuGenerics.TimeEpochDiff(epochs)
	},
}

func init() {
	TimeCmd.AddCommand(timeNowCmd)
	TimeCmd.AddCommand(timeParseCmd)
	TimeCmd.AddCommand(timeUntilCmd)
	TimeCmd.AddCommand(timeDiffCmd)
}
