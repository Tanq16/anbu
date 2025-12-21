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
	Short:   "Start a markdown viewer web server",
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
