package cmd

import (
	"github.com/spf13/cobra"
	anbuTime "github.com/tanq16/anbu/internal/time"
)

var timeCmdFlags struct {
	epochs      []int64
	action      string
	parseAction string
	timeStr     string
}

var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "time related function: use `now`, `purple`, `diff` (calculate epoch diff), `parse` (ingest a time str & print)",
	Run: func(cmd *cobra.Command, args []string) {
		switch timeCmdFlags.action {
		case "now":
			anbuTime.Current()
		case "purple":
			anbuTime.Purple()
		case "diff":
			anbuTime.EpochDiff(timeCmdFlags.epochs)
		case "parse":
			anbuTime.Parse(timeCmdFlags.timeStr, timeCmdFlags.parseAction)
		default:
			anbuTime.Current()
		}
	},
}

func init() {
	timeCmd.Flags().StringVarP(&timeCmdFlags.action, "action", "a", "", "Action to perform: now, purple, diff, parse")
	timeCmd.Flags().Int64SliceVarP(&timeCmdFlags.epochs, "epochs", "e", []int64{}, "Epochs to calculate difference between")
	timeCmd.Flags().StringVarP(&timeCmdFlags.parseAction, "parse-action", "p", "normal", "Parse action: normal, purple")
	timeCmd.Flags().StringVarP(&timeCmdFlags.timeStr, "time-str", "t", "", "Time string to parse")

	rootCmd.AddCommand(timeCmd)
}
