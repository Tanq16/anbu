package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/utils"
)

var tableTestCmd = &cobra.Command{
	Use:   "testcommand",
	Short: "command to test functions",
	Run: func(cmd *cobra.Command, args []string) {
		table := utils.MarkdownTable{
			Headers: []string{"Component", "Status", "Version", "Last Updated"},
			Rows: [][]string{
				{"Operating System", "Running", "Linux 6.5", "2025-03-15"},
				{"Database Wrap around the name", "Active", "PostgreSQL 16.2", "2025-04-01"},
				{"Web Server", "Running", "Nginx 1.28.0", "2025-03-22"},
				{"API Gateway", "Warning", "Kong 3.4.2", "2025-02-10"},
				{"Cache", "Inactive", "Redis 8.0.1", "2025-01-30"},
			},
		}
		err := table.OutMDPrint(true)
		if err != nil {
			utils.OutError("Failed to print table: " + err.Error())
		} else {
			utils.OutSuccess("Table printed successfully")
		}
		// err = table.OutMDFile("test.md")
		// if err != nil {
		// 	utils.OutError("Failed to write table to file: " + err.Error())
		// } else {
		// 	utils.OutSuccess("Table written to test.md")
		// }

		utils.OutWarning("Warning: This is a test warning message")
		utils.OutError("Error: This is a test error message")
		utils.OutSuccess("Success: This is a test success message")
		utils.OutInfo("Info: This is a test info message")
		utils.OutDebug("Debug: This is a test debug message")
		fmt.Println("Test message with no color")
		fmt.Println("Test message with no color")
		fmt.Println("Test message with no color")
		fmt.Println("Test message with no color")
		utils.OutClearLines(3)

		logger := utils.GetLogger("test")
		logger.Debug().Msg("This is a debug message")
		logger.Info().Msg("This is an info message")
		logger.Warn().Msg("This is a warning message")
		logger.Error().Msg("This is an error message")
		logger.Fatal().Msg("This is a fatal message")
	},
}

func init() {
	rootCmd.AddCommand(tableTestCmd)
}
