package genericsCmd

import (
	"fmt"

	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	"github.com/tanq16/anbu/utils"
)

var loopCmdFlagPadding int
var loopCmdFlagCommand string
var loopCmdFlagRange []int

var LoopCmd = &cobra.Command{
	Use:   "loop",
	Short: "execute a command for each number range in a range",
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("loopcmd")
		if len(args) < 2 {
			logger.Fatal().Msg("Missing count or command")
		} else {
			var err error
			loopCmdFlagRange, err = anbuGenerics.LoopProcessRange(args[0])
			if err != nil {
				logger.Fatal().Err(err).Msg("Not a valid count")
			}
			loopCmdFlagCommand = args[1]
		}
		err := anbuGenerics.LoopProcessCommands(loopCmdFlagRange, loopCmdFlagCommand, loopCmdFlagPadding)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to execute linear operation")
		}
		fmt.Println(utils.OutSuccess("Linear operation completed successfully"))
	},
}

func init() {
	LoopCmd.Flags().IntVarP(&loopCmdFlagPadding, "padding", "p", 0, "padding for the number")
}
