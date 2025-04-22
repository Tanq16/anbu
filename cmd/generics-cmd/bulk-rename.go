package genericsCmd

import (
	"github.com/spf13/cobra"
)

var bulkRenameFlags struct {
	renameDirectories bool
}

var BulkRenameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Bulk rename files/directories using regex pattern and replacement as args",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// pattern := args[0]
		// replacement := args[1]
		// err := anbuGenerics.BulkRename(pattern, replacement, bulkRenameFlags.renameDirectories)
		// if err != nil {
		// 	logger.Fatal().Err(err).Msg("Bulk rename operation failed")
		// }
	},
}

func init() {
	BulkRenameCmd.Flags().BoolVarP(&bulkRenameFlags.renameDirectories, "directories", "d", false, "Rename directories instead of files")
}
