package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

var bulkRenameFlags struct {
	renameDirectories bool
	dryRun            bool
}

var BulkRenameCmd = &cobra.Command{
	Use:     "rename <pattern> <replacement>",
	Aliases: []string{},
	Short:   "Bulk rename files or directories using regex patterns",
	Long: `Rename multiple files or directories in a single operation using regex patterns.
Examples:
  anbu rename 'old_(.*)' 'new_\1'                 # Rename files matching regex pattern
  anbu rename -d 'old_(.*)' 'new_\1'              # Rename directories instead of files
  anbu rename '(.*)\.(.*)' '\1_backup.\2'         # Add _backup before extension`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		anbuGenerics.BulkRename(args[0], args[1], bulkRenameFlags.renameDirectories, bulkRenameFlags.dryRun)
	},
}

func init() {
	BulkRenameCmd.Flags().BoolVarP(&bulkRenameFlags.renameDirectories, "directories", "d", false, "Rename directories instead of files")
	BulkRenameCmd.Flags().BoolVarP(&bulkRenameFlags.dryRun, "dry-run", "r", false, "Simulate the rename operation without making changes")
}
