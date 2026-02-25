package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/internal/utils"
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
		if err := anbuGenerics.FindDuplicates(duplicatesFlags.recursive, duplicatesFlags.delete); err != nil {
			u.PrintFatal("find duplicates failed", err)
		}
	},
}

func init() {
	DuplicatesCmd.Flags().BoolVarP(&duplicatesFlags.recursive, "recursive", "r", false, "Search recursively in subdirectories")
	DuplicatesCmd.Flags().BoolVar(&duplicatesFlags.delete, "delete", false, "Delete duplicate files, keeping only the first copy in each set")
}
