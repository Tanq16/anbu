package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
)

var loopCmdFlagPadding int

var LoopCmd = &cobra.Command{
	Use:   "loop",
	Short: "execute a command for each number range in a range",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			u.PrintError("Missing count or command")
			return
		}
		anbuGenerics.LoopProcessCommands(args[0], args[1], loopCmdFlagPadding)
	},
}

func init() {
	LoopCmd.Flags().IntVarP(&loopCmdFlagPadding, "padding", "p", 0, "padding for the number")
}
