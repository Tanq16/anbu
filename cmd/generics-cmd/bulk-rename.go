package genericsCmd

import (
	"fmt"

	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
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
  anbu rename --directories 'old_(.*)' 'new_\1'   # Rename directories instead of files
  anbu rename '(.*)\.(.*)' '\1_backup.\2'         # Add _backup before extension`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		results, err := anbuGenerics.BulkRename(args[0], args[1], bulkRenameFlags.renameDirectories, bulkRenameFlags.dryRun)
		if err != nil {
			u.PrintFatal("bulk rename failed", err)
		}
		kind := "files"
		if bulkRenameFlags.renameDirectories {
			kind = "directories"
		}
		renamed := 0
		for _, result := range results {
			if result.Err != nil {
				u.PrintError(fmt.Sprintf("Failed to rename %s to %s", result.Old, result.New), result.Err)
				continue
			}
			label := "Renamed:"
			if bulkRenameFlags.dryRun {
				label = "Dry Run: Renaming"
			}
			u.PrintGeneric(fmt.Sprintf("%s %s %s %s", label, u.FDebug(result.Old), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(result.New)))
			renamed++
		}
		u.LineBreak()
		if renamed == 0 {
			u.PrintWarn("no items were renamed", nil)
			return
		}
		u.PrintGeneric(fmt.Sprintf("%s %s", u.FDebug("Operation completed:"), u.FSuccess(fmt.Sprintf("%d %s", renamed, kind))))
	},
}

func init() {
	BulkRenameCmd.Flags().BoolVar(&bulkRenameFlags.renameDirectories, "directories", false, "Rename directories instead of files")
	BulkRenameCmd.Flags().BoolVar(&bulkRenameFlags.dryRun, "dry-run", false, "Simulate the rename operation without making changes")
}
