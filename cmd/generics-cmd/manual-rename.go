package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

var manualRenameFlags struct {
	includeDir       bool
	hidden           bool
	includeExtension bool
}

var ManualRenameCmd = &cobra.Command{
	Use:     "manual-rename",
	Aliases: []string{"mrename"},
	Short:   "Manually rename files and directories one by one",
	Long: `Interactively rename files and directories one by one.
For each file or directory, you will be prompted to enter a new name.
Press Enter without typing to skip renaming an item.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		anbuGenerics.ManualRename(manualRenameFlags.includeDir, manualRenameFlags.hidden, manualRenameFlags.includeExtension)
	},
}

func init() {
	ManualRenameCmd.Flags().BoolVarP(&manualRenameFlags.includeDir, "include-dir", "d", false, "Include directories in the rename operation")
	ManualRenameCmd.Flags().BoolVarP(&manualRenameFlags.hidden, "hidden", "H", false, "Include hidden files and directories")
	ManualRenameCmd.Flags().BoolVarP(&manualRenameFlags.includeExtension, "include-extension", "x", false, "Allow changing file extension")
}
