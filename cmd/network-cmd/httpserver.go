package networkCmd

import (
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
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		server := anbuNetwork.NewHTTPServer(&anbuNetwork.HTTPServerOptions{
			ListenAddress: httpServerFlags.listenAddress,
			EnableUpload:  httpServerFlags.enableUpload,
			EnableTLS:     httpServerFlags.enableTLS,
		})
		if err := server.Setup(); err != nil {
			return err
		}
		defer server.Stop()
		return server.Run()
	},
}

func init() {
	HTTPServerCmd.Flags().StringVarP(&httpServerFlags.listenAddress, "listen", "l", "0.0.0.0:8080", "Address and port to listen on")
	HTTPServerCmd.Flags().BoolVar(&httpServerFlags.enableUpload, "upload", false, "Enable file uploads via PUT requests")
	HTTPServerCmd.Flags().BoolVar(&httpServerFlags.enableTLS, "tls", false, "Enable HTTPS with a self-signed certificate")
}
