package networkCmd

import (
	"github.com/spf13/cobra"
	anbuNetwork "github.com/tanq16/anbu/internal/network"
	"github.com/tanq16/anbu/utils"
)

var httpServerFlags struct {
	listenAddress string
	enableUpload  bool
	enableTLS     bool
	domain        string
}

var HTTPServerCmd = &cobra.Command{
	Use:   "httpserver",
	Short: "Start a simple HTTP/HTTPS file server",
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("httpserver")
		options := &anbuNetwork.HTTPServerOptions{
			ListenAddress: httpServerFlags.listenAddress,
			EnableUpload:  httpServerFlags.enableUpload,
			EnableTLS:     httpServerFlags.enableTLS,
			Domain:        httpServerFlags.domain,
		}
		server := &anbuNetwork.HTTPServer{
			Options: options,
		}
		err := server.Start()
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
		defer server.Stop()
	},
}

func init() {
	HTTPServerCmd.Flags().StringVarP(&httpServerFlags.listenAddress, "listen", "l", "localhost:8000", "Address:Port to listen on")
	HTTPServerCmd.Flags().BoolVarP(&httpServerFlags.enableUpload, "upload", "u", false, "Enable file uploads via PUT")
	HTTPServerCmd.Flags().BoolVarP(&httpServerFlags.enableTLS, "tls", "t", false, "Enable HTTPS (TLS)")
	HTTPServerCmd.Flags().StringVar(&httpServerFlags.domain, "domain", "localhost", "Domain for self-signed certificate")
}
