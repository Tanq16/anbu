package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

var sedFlags struct {
	dryRun bool
}

var SedCmd = &cobra.Command{
	Use:     "sed <pattern> <replacement> <path>",
	Aliases: []string{},
	Short:   "Apply regex substitution to file content",
	Long: `Replace text patterns in single files or entire directories using regex patterns.

Examples:
  anbu sed 'old_(.*)' 'new_\1' path/to/file.txt  # Replace text in file
  anbu sed 'old_(.*)' 'new_\1' path/to/dir       # Replace text in all files in directory
  anbu sed 'old_(.*)' 'new_\1' path/to/dir -r    # Perform a dry-run without applying changes`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		anbuGenerics.Sed(args[0], args[1], args[2], sedFlags.dryRun)
	},
}

func init() {
	SedCmd.Flags().BoolVarP(&sedFlags.dryRun, "dry-run", "r", false, "Show file content with substitutions without writing")
}
