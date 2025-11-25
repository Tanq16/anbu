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
	Short:   "Apply regex substitution to file(s)",
	Long: `Applies regex pattern matching and replacement to file content.
The first expression is a regex pattern, the second is a replacement string.
If path is a file, applies to that file. If path is a directory, applies to all files within.

- Use \1, \2, etc. in the <replacement> string to refer to capture groups from the <pattern>.
- The pattern is standard Go regex. Remember to quote arguments to prevent shell expansion.
- Substitutions are applied globally to each file.

Examples:
  # Replace all occurrences of 'old' with 'new' in a file
  anbu sed 'old' 'new' file.txt

  # Replace email patterns with masked version
  anbu sed '([a-z]+)@([a-z]+)\.com' '\1@***.com' file.txt

  # Simulate substitution without making changes
  anbu sed 'foo' 'bar' ./directory -r`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		anbuGenerics.Sed(args[0], args[1], args[2], sedFlags.dryRun)
	},
}

func init() {
	SedCmd.Flags().BoolVarP(&sedFlags.dryRun, "dry-run", "r", false, "Show file content with substitutions without writing")
}
