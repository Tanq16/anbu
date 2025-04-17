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
	Short: "print times with args `now`, `purple`, `diff`, `parse`",
	Long: `Arguments:
- now: print the current time in various formats
- purple: print the current time in purple team format (includes public ip)
- diff: print the difference between two epochs
- parse: parse a time string across various formats and print as table
`,
	Args: cobra.MaximumNArgs(1),
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
		case "diff":
			anbuGenerics.TimeEpochDiff(timeCmdFlags.epochs)
		case "parse":
			anbuGenerics.TimeParse(timeCmdFlags.timeStr, timeCmdFlags.parseAction)
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
