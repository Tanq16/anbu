package genericsCmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
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
		result, err := anbuGenerics.FindDuplicates(duplicatesFlags.recursive, duplicatesFlags.delete)
		if err != nil {
			u.PrintFatal("find duplicates failed", err)
		}
		if len(result.Hashed) == 0 && len(result.Unhashed) == 0 {
			u.PrintInfo("No duplicate files found")
			return
		}
		if len(result.Hashed) > 0 {
			printDuplicateTable(result.Hashed, 1)
		}
		if len(result.Unhashed) > 0 {
			u.LineBreak()
			u.PrintWarn("Unhashed duplicates due to huge size:", nil)
			printDuplicateTable(result.Unhashed, len(result.Hashed)+1)
		}
		for _, path := range result.Deleted {
			u.PrintGeneric(fmt.Sprintf("Deleted: %s", u.FSuccess(path)))
		}
		for _, failure := range result.DeleteErrs {
			u.PrintError(fmt.Sprintf("Failed to delete %s", failure.Path), failure.Err)
		}
	},
}

func printDuplicateTable(sets []anbuGenerics.DuplicateSet, startID int) {
	table := u.NewTable([]string{"Set ID", "Files"})
	for i, set := range sets {
		table.Rows = append(table.Rows, []string{fmt.Sprintf("%d", startID+i), strings.Join(set.Files, ", ")})
	}
	table.PrintTable()
}

func init() {
	DuplicatesCmd.Flags().BoolVar(&duplicatesFlags.recursive, "recursive", false, "Search recursively in subdirectories")
	DuplicatesCmd.Flags().BoolVar(&duplicatesFlags.delete, "delete", false, "Delete duplicate files, keeping only the first copy in each set")
}
