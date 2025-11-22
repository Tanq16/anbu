package genericsCmd

import (
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

var StashCmd = &cobra.Command{
	Use:   "stash",
	Short: "Manage a persistent clipboard for files, folders, and text snippets",
	Long: `Stash provides a persistent storage for files, folders, and text snippets.

Subcommands:
  fs <path>      Stash a file or folder (removes original)
  text <name>    Stash text from stdin
  list           List all stashed entries
  apply <id>     Apply a stash without removing it
  pop <id>       Apply a stash and remove it
  clear <id>     Remove a stash without applying it`,
}

var stashFSCmd = &cobra.Command{
	Use:   "fs <path>",
	Short: "Stash a file or folder",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.StashFS(args[0]); err != nil {
			log.Fatal().Err(err).Msg("failed to stash")
		}
	},
}

var stashTextCmd = &cobra.Command{
	Use:   "text <name>",
	Short: "Stash text from stdin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.StashText(args[0]); err != nil {
			log.Fatal().Err(err).Msg("failed to stash text")
		}
	},
}

var stashListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stashed entries",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.StashList(); err != nil {
			log.Fatal().Err(err).Msg("failed to list stashes")
		}
	},
}

var stashApplyCmd = &cobra.Command{
	Use:   "apply <id>",
	Short: "Apply a stash without removing it",
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
	Short: "Apply a stash and remove it",
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
	Short: "Remove a stash without applying it",
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
