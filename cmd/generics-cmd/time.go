package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

var timeCmdFlags struct {
	epochs      []int64
	action      string
	parseAction string
	timeStr     string
}

var TimeCmd = &cobra.Command{
	Use:   "time",
	Short: "time related function: use `now`, `purple`, `diff` (calculate epoch diff), `parse` (ingest a time str & print)",
	Run: func(cmd *cobra.Command, args []string) {
		switch timeCmdFlags.action {
		case "now":
			anbuGenerics.Current()
		case "purple":
			anbuGenerics.Purple()
		case "diff":
			anbuGenerics.EpochDiff(timeCmdFlags.epochs)
		case "parse":
			anbuGenerics.Parse(timeCmdFlags.timeStr, timeCmdFlags.parseAction)
		default:
			anbuGenerics.Current()
		}
	},
}

func init() {
	TimeCmd.Flags().StringVarP(&timeCmdFlags.action, "action", "a", "", "Action to perform: now, purple, diff, parse")
	TimeCmd.Flags().Int64SliceVarP(&timeCmdFlags.epochs, "epochs", "e", []int64{}, "Epochs to calculate difference between")
	TimeCmd.Flags().StringVarP(&timeCmdFlags.parseAction, "parse-action", "p", "normal", "Parse action: normal, purple")
	TimeCmd.Flags().StringVarP(&timeCmdFlags.timeStr, "time-str", "t", "", "Time string to parse")
}
