package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

var duplicatesFlags struct {
	recursive bool
	delete    bool
}

var DuplicatesCmd = &cobra.Command{
	Use:     "duplicates",
	Aliases: []string{"dup"},
	Short:   "Find duplicate files by content with optional recursive search",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		anbuGenerics.FindDuplicates(duplicatesFlags.recursive, duplicatesFlags.delete)
	},
}

func init() {
	DuplicatesCmd.Flags().BoolVarP(&duplicatesFlags.recursive, "recursive", "r", false, "Search recursively in subdirectories")
	DuplicatesCmd.Flags().BoolVar(&duplicatesFlags.delete, "delete", false, "Delete duplicate files, keeping only the first copy in each set")
}
