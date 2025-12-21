package networkCmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	anbuNetwork "github.com/tanq16/anbu/internal/network"
)

var httpServerFlags struct {
	listenAddress string
	enableUpload  bool
	enableTLS     bool
}

var HTTPServerCmd = &cobra.Command{
	Use:   "http-server",
	Short: "Start a simple HTTP/HTTPS file server with optional file uploads",
	Run: func(cmd *cobra.Command, args []string) {
		options := &anbuNetwork.HTTPServerOptions{
			ListenAddress: httpServerFlags.listenAddress,
			EnableUpload:  httpServerFlags.enableUpload,
			EnableTLS:     httpServerFlags.enableTLS,
		}
		server := &anbuNetwork.HTTPServer{
			Options: options,
		}
		err := server.Start()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
		defer server.Stop()
	},
}

func init() {
	HTTPServerCmd.Flags().StringVarP(&httpServerFlags.listenAddress, "listen", "l", "0.0.0.0:8080", "Address and port to listen on")
	HTTPServerCmd.Flags().BoolVarP(&httpServerFlags.enableUpload, "upload", "u", false, "Enable file uploads via PUT requests")
	HTTPServerCmd.Flags().BoolVarP(&httpServerFlags.enableTLS, "tls", "t", false, "Enable HTTPS with a self-signed certificate")
}
