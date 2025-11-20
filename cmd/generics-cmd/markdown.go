package genericsCmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

var markdownFlags struct {
	listenAddress string
}

var MarkdownCmd = &cobra.Command{
	Use:     "markdown",
	Aliases: []string{"md"},
	Short:   "Start a markdown viewer web server for the current directory",
	Long: `Starts a web server that displays all files in the current directory
in a sidebar and renders markdown files.

Examples:
  # Start markdown viewer on default address (0.0.0.0:8080)
  anbu markdown

  # Start on a specific address and port
  anbu markdown -l 127.0.0.1:9090
  anbu md -l :3000`,
	Run: func(cmd *cobra.Command, args []string) {
		err := anbuGenerics.StartMarkdownServer(markdownFlags.listenAddress)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to start markdown viewer")
		}
	},
}

func init() {
	MarkdownCmd.Flags().StringVarP(&markdownFlags.listenAddress, "listen", "l", "0.0.0.0:8080", "Address and port to listen on")
}
