package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	anbuLoopCmd "github.com/tanq16/anbu/internal/loopcmd"
	"github.com/tanq16/anbu/utils"
)

var loopCmdFlagPadding int
var loopCmdFlagCommand string
var loopCmdFlagRange []int

var loopCmd = &cobra.Command{
	Use:   "loop",
	Short: "execute a command for each number range in a range",
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("loopcmd")
		if len(args) < 2 {
			logger.Fatal().Msg("Missing count or command")
		} else {
			var err error
			loopCmdFlagRange, err = anbuLoopCmd.ProcessRange(args[0])
			if err != nil {
				logger.Fatal().Err(err).Msg("Not a valid count")
			}
			loopCmdFlagCommand = args[1]
		}
		err := anbuLoopCmd.ProcessCommands(loopCmdFlagRange, loopCmdFlagCommand, loopCmdFlagPadding)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to execute linear operation")
		}
		fmt.Println(utils.OutSuccess("Linear operation completed successfully"))
	},
}

func init() {
	loopCmd.Flags().IntVarP(&loopCmdFlagPadding, "padding", "p", 0, "padding for the number")
	rootCmd.AddCommand(loopCmd)
}
