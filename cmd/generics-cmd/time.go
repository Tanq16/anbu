package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
)

var timeDiffFlags struct {
	epochs []int64
}

var TimeCmd = &cobra.Command{
	Use:     "time",
	Aliases: []string{"t"},
	Short:   "Show times in common formats, parse a string, or diff epochs",
	Args:    cobra.NoArgs,
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
	Use:   "diff",
	Short: "Print the difference between Unix epochs",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.TimeEpochDiff(timeDiffFlags.epochs); err != nil {
			u.PrintFatal("no epochs provided", err)
		}
	},
}

func init() {
	TimeCmd.AddCommand(timeParseCmd)
	TimeCmd.AddCommand(timeUntilCmd)
	TimeCmd.AddCommand(timeDiffCmd)

	timeDiffCmd.Flags().Int64SliceVarP(&timeDiffFlags.epochs, "epochs", "e", []int64{}, "Unix epochs (repeatable)")
}
