package genericsCmd

import (
	"fmt"

	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

type parseAction string

func (p *parseAction) String() string { return string(*p) }
func (p *parseAction) Type() string   { return "normal|purple|diff" }

func (p *parseAction) Set(v string) error {
	switch v {
	case "normal", "purple", "diff":
		*p = parseAction(v)
		return nil
	}
	return fmt.Errorf("must be one of normal, purple, diff")
}

var timeCmdFlags struct {
	epochs      []int64
	parseAction parseAction
	timeStr     string
}

var TimeCmd = &cobra.Command{
	Use:     "time [now|purple|iso|diff|parse|until]",
	Aliases: []string{"t"},
	Short:   "Display and analyze time in various formats and perform epoch diffs, time parsing, and time remaining calculations",
	Args:    cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{
		"now",
		"purple",
		"iso",
		"diff",
		"parse",
		"until",
	},
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
			anbuGenerics.TimeParse(timeCmdFlags.timeStr, string(timeCmdFlags.parseAction))
		case "until":
			anbuGenerics.TimeParse(timeCmdFlags.timeStr, "diff")
		}
	},
}

func init() {
	timeCmdFlags.parseAction = "normal"
	TimeCmd.Flags().Int64SliceVarP(&timeCmdFlags.epochs, "epochs", "e", []int64{}, "Epochs to calculate difference between")
	TimeCmd.Flags().VarP(&timeCmdFlags.parseAction, "parse-action", "p", "Parse action")
	TimeCmd.Flags().StringVarP(&timeCmdFlags.timeStr, "time-str", "t", "", "Time string to parse")
}
