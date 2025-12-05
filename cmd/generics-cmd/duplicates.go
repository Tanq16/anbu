package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

var duplicatesFlags struct {
	recursive bool
}

var DuplicatesCmd = &cobra.Command{
	Use:     "duplicates",
	Aliases: []string{"dup"},
	Short:   "Find duplicate files in the current directory",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		anbuGenerics.FindDuplicates(duplicatesFlags.recursive)
	},
}

func init() {
	DuplicatesCmd.Flags().BoolVarP(&duplicatesFlags.recursive, "recursive", "r", false, "Search recursively in subdirectories")
}
