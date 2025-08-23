package networkCmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	anbuNetwork "github.com/tanq16/anbu/internal/network"
	u "github.com/tanq16/anbu/utils"
)

var httpServerFlags struct {
	listenAddress string
	enableUpload  bool
	enableTLS     bool
	domain        string
}

var HTTPServerCmd = &cobra.Command{
	Use:   "http-server",
	Short: "Start a simple HTTP/HTTPS file server for the current directory",
	Long: `Serves files from the current directory over HTTP or HTTPS.
It can also be configured to accept file uploads via PUT requests.
When TLS is enabled, it generates a self-signed certificate.

Examples:
  # Serve current directory on http://0.0.0.0:8000
  anbu http-server

  # Serve on a different address and port
  anbu http-server -l 127.0.0.1:9090

  # Enable file uploads (e.g., curl -T file.txt http://localhost:8000/file.txt)
  anbu http-server -u

  # Serve over HTTPS with a self-signed certificate
  anbu http-server -t

  # Serve HTTPS for a specific domain in the cert
  anbu http-server -t --domain my.local.dev`,
	Run: func(cmd *cobra.Command, args []string) {
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
			u.PrintError(fmt.Sprintf("Failed to start HTTP server: %v", err))
			os.Exit(1)
		}
		defer server.Stop()
	},
}

func init() {
	HTTPServerCmd.Flags().StringVarP(&httpServerFlags.listenAddress, "listen", "l", "0.0.0.0:8000", "Address and port to listen on")
	HTTPServerCmd.Flags().BoolVarP(&httpServerFlags.enableUpload, "upload", "u", false, "Enable file uploads via PUT requests")
	HTTPServerCmd.Flags().BoolVarP(&httpServerFlags.enableTLS, "tls", "t", false, "Enable HTTPS with a self-signed certificate")
	HTTPServerCmd.Flags().StringVar(&httpServerFlags.domain, "domain", "localhost", "Domain to use for the self-signed certificate")
}
