package cmd

// import (
// 	"fmt"
// 	"os"
// 	"strconv"

// 	"github.com/spf13/cobra"
// 	"github.com/tanq16/anbu/internal/fileutil"
// 	"github.com/tanq16/anbu/utils"
// )

// var forLinearCmd = &cobra.Command{
// 	Use:   "forlinear",
// 	Short: "Execute a command for each number in a range",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		logger := utils.GetLogger("fileutil")

// 		if len(args) < 2 {
// 			if fileUtilFlags.count == 0 || fileUtilFlags.command == "" {
// 				logger.Error().Msg("Missing count or command")
// 				fmt.Println(utils.OutError("Error: provide count and command"))
// 				cmd.Help()
// 				os.Exit(1)
// 			}
// 		} else {
// 			// Parse from args if flags not provided
// 			var err error
// 			fileUtilFlags.count, err = strconv.Atoi(args[0])
// 			if err != nil {
// 				logger.Error().Err(err).Msg("Invalid count")
// 				fmt.Println(utils.OutError("Error: count must be a number"))
// 				os.Exit(1)
// 			}
// 			fileUtilFlags.command = args[1]
// 		}

// 		err := fileutil.ProcessLinearOperation(fileUtilFlags.count, fileUtilFlags.command)
// 		if err != nil {
// 			logger.Error().Err(err).Msg("Failed to execute linear operation")
// 			fmt.Println(utils.OutError("Error: " + err.Error()))
// 			os.Exit(1)
// 		}

// 		fmt.Println(utils.OutSuccess("Linear operation completed successfully"))
// 	},
// }

// func init() {
// 	forLinearCmd.Flags().IntVarP(&fileUtilFlags.count, "count", "c", 0, "Count for the loop")
// 	forLinearCmd.Flags().StringVarP(&fileUtilFlags.command, "command", "x", "", "Command to execute")
// 	fileUtilCmd.AddCommand(forLinearCmd)
// }
