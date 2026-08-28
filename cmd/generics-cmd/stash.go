package genericsCmd

import (
	"errors"
	"strconv"

	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
)

var stashTextFlags struct {
	text string
}

var StashCmd = &cobra.Command{
	Use:   "stash",
	Short: "Manage a persistent clipboard for files, folders, and text",
}

var stashFSCmd = &cobra.Command{
	Use:   "fs <path>",
	Short: "Stash a file or folder keeping the original in its location",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.StashFS(args[0]); err != nil {
			u.PrintFatal("failed to stash", err)
		}
	},
}

var stashTextCmd = &cobra.Command{
	Use:   "text <name>",
	Short: "Stash text from --text or an interactive editor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := stashTextFlags.text
		if text == "" {
			var err error
			text, err = u.PromptTextArea("Enter text to stash (name: "+args[0]+"):", "")
			if errors.Is(err, u.ErrNoTerminal) {
				u.PrintFatal("stash text needs --text when there is no interactive terminal", nil)
			}
			if err != nil {
				u.PrintFatal("failed to stash text", err)
			}
		}
		if text == "" {
			u.PrintFatal("stash text needs --text when there is no interactive terminal", nil)
		}
		if err := anbuGenerics.StashText(args[0], text); err != nil {
			u.PrintFatal("failed to stash text", err)
		}
	},
}

var stashListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stashed entries with their IDs and types",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.StashList(); err != nil {
			u.PrintFatal("failed to list stashes", err)
		}
	},
}

var stashApplyCmd = &cobra.Command{
	Use:   "apply <id>",
	Short: "Apply a stash without removing it (text prints to stdout, files/folders extracted to current directory)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			u.PrintFatal("invalid stash ID", err)
		}
		if err := anbuGenerics.StashApply(id); err != nil {
			u.PrintFatal("failed to apply stash", err)
		}
	},
}

var stashPopCmd = &cobra.Command{
	Use:   "pop <id>",
	Short: "Apply a stash and remove it (text prints to stdout, files/folders extracted to current directory)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			u.PrintFatal("invalid stash ID", err)
		}
		if err := anbuGenerics.StashPop(id); err != nil {
			u.PrintFatal("failed to pop stash", err)
		}
	},
}

var stashClearCmd = &cobra.Command{
	Use:   "clear <id>",
	Short: "Remove a stash without applying or popping it",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			u.PrintFatal("invalid stash ID", err)
		}
		if err := anbuGenerics.StashClear(id); err != nil {
			u.PrintFatal("failed to clear stash", err)
		}
	},
}

func init() {
	stashTextCmd.Flags().StringVar(&stashTextFlags.text, "text", "", "Text to stash")

	StashCmd.AddCommand(stashFSCmd)
	StashCmd.AddCommand(stashTextCmd)
	StashCmd.AddCommand(stashListCmd)
	StashCmd.AddCommand(stashApplyCmd)
	StashCmd.AddCommand(stashPopCmd)
	StashCmd.AddCommand(stashClearCmd)
}
