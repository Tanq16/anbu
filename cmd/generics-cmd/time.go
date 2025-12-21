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
	Use:     "time",
	Aliases: []string{"t"},
	Short:   "Display and analyze time in various formats and perform epoch diffs, time parsing, and time remaining calculations",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			anbuGenerics.TimeCurrent()
			return
		}
		switch args[0] {
		case "now":
			anbuGenerics.TimeCurrent()
		case "purple":
			anbuGenerics.TimePurple()
		case "iso":
			anbuGenerics.TimeISO()
		case "diff":
			anbuGenerics.TimeEpochDiff(timeCmdFlags.epochs)
		case "parse":
			anbuGenerics.TimeParse(timeCmdFlags.timeStr, timeCmdFlags.parseAction)
		case "until":
			anbuGenerics.TimeParse(timeCmdFlags.timeStr, "diff")
		default:
			anbuGenerics.TimeCurrent()
		}
	},
}

func init() {
	TimeCmd.Flags().Int64SliceVarP(&timeCmdFlags.epochs, "epochs", "e", []int64{}, "Epochs to calculate difference between")
	TimeCmd.Flags().StringVarP(&timeCmdFlags.parseAction, "parse-action", "p", "normal", "Parse action: normal, purple")
	TimeCmd.Flags().StringVarP(&timeCmdFlags.timeStr, "time-str", "t", "", "Time string to parse")
}
