package genericsCmd

import (
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

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
			log.Fatal().Err(err).Msg("failed to stash")
		}
	},
}

var stashTextCmd = &cobra.Command{
	Use:   "text <name>",
	Short: "Stash text from stdin and stash it with a given name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.StashText(args[0]); err != nil {
			log.Fatal().Err(err).Msg("failed to stash text")
		}
	},
}

var stashListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stashed entries with their IDs and types",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.StashList(); err != nil {
			log.Fatal().Err(err).Msg("failed to list stashes")
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
			log.Fatal().Msg("invalid stash ID")
		}
		if err := anbuGenerics.StashApply(id); err != nil {
			log.Fatal().Err(err).Msg("failed to apply stash")
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
			log.Fatal().Msg("invalid stash ID")
		}
		if err := anbuGenerics.StashPop(id); err != nil {
			log.Fatal().Err(err).Msg("failed to pop stash")
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
			log.Fatal().Msg("invalid stash ID")
		}
		if err := anbuGenerics.StashClear(id); err != nil {
			log.Fatal().Err(err).Msg("failed to clear stash")
		}
	},
}

func init() {
	StashCmd.AddCommand(stashFSCmd)
	StashCmd.AddCommand(stashTextCmd)
	StashCmd.AddCommand(stashListCmd)
	StashCmd.AddCommand(stashApplyCmd)
	StashCmd.AddCommand(stashPopCmd)
	StashCmd.AddCommand(stashClearCmd)
}
