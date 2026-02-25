package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/internal/utils"
)

var manualRenameFlags struct {
	includeDir       bool
	hidden           bool
	includeExtension bool
}

var ManualRenameCmd = &cobra.Command{
	Use:     "manual-rename",
	Aliases: []string{"mrename"},
	Short:   "Interactively rename files and directories one by one, optionally including directories, hidden files, and extensions",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.ManualRename(manualRenameFlags.includeDir, manualRenameFlags.hidden, manualRenameFlags.includeExtension); err != nil {
			u.PrintFatal("manual rename failed", err)
		}
	},
}

func init() {
	ManualRenameCmd.Flags().BoolVarP(&manualRenameFlags.includeDir, "include-dir", "d", false, "Include directories in the rename operation")
	ManualRenameCmd.Flags().BoolVarP(&manualRenameFlags.hidden, "hidden", "H", false, "Include hidden files and directories")
	ManualRenameCmd.Flags().BoolVarP(&manualRenameFlags.includeExtension, "include-extension", "x", false, "Allow changing file extension")
}
