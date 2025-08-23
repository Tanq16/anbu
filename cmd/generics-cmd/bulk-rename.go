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
	Short:   "Bulk rename files or directories using regex",
	Long: `Renames multiple files or directories in the current folder based on a regex pattern.
It uses capture groups from the pattern in the replacement string.

- Use \\1, \\2, etc. in the <replacement> string to refer to capture groups from the <pattern>.
- The pattern is standard Go regex. Remember to quote arguments to prevent shell expansion.

Examples:
  # Add a prefix 'new_' to files starting with 'prefix_'
  anbu rename 'prefix_(.*)' 'new_\\1'

  # Do the same for directories instead of files
  anbu rename 'old_(.*)' 'new_\\1' -d

  # Add '_backup' before the file extension (e.g., file.txt -> file_backup.txt)
  anbu rename '(.*)\\.(.*)' '\\1_backup.\\2'

  # Simulate a rename operation without making changes
  anbu rename 'image-(\\d+).jpg' 'IMG_\\1.jpeg' -r`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		anbuGenerics.BulkRename(args[0], args[1], bulkRenameFlags.renameDirectories, bulkRenameFlags.dryRun)
	},
}

func init() {
	BulkRenameCmd.Flags().BoolVarP(&bulkRenameFlags.renameDirectories, "directories", "d", false, "Rename directories instead of files")
	BulkRenameCmd.Flags().BoolVarP(&bulkRenameFlags.dryRun, "dry-run", "r", false, "Simulate the rename operation without making changes")
}
