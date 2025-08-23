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
	Use:     "rename",
	Aliases: []string{},
	Short:   "Bulk rename files/directories using regex pattern and replacement as args",
	Long: `Examples:
- s
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		anbuGenerics.BulkRename(args[0], args[1], bulkRenameFlags.renameDirectories, bulkRenameFlags.dryRun)
	},
}

func init() {
	BulkRenameCmd.Flags().BoolVarP(&bulkRenameFlags.renameDirectories, "directories", "d", false, "Rename directories instead of files")
	BulkRenameCmd.Flags().BoolVarP(&bulkRenameFlags.dryRun, "dry-run", "r", false, "Simulate the rename operation without actually renaming files")
}
