package networkCmd

import (
	"fmt"

	"github.com/spf13/cobra"
	anbuNetwork "github.com/tanq16/anbu/internal/network"
	u "github.com/tanq16/anbu/utils"
)

var httpServerFlags struct {
	listenAddress string
	enableUpload  bool
}

var HTTPServerCmd = &cobra.Command{
	Use:   "http-server",
	Short: "Start a simple HTTP file server with optional file uploads",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		server := anbuNetwork.NewHTTPServer(&anbuNetwork.HTTPServerOptions{
			ListenAddress: httpServerFlags.listenAddress,
			EnableUpload:  httpServerFlags.enableUpload,
			OnRequest: func(remote, method, path string) {
				u.PrintStream(fmt.Sprintf("%s %s %s", remote, method, path))
			},
			OnInfo:  u.PrintInfo,
			OnError: u.PrintError,
		})
		if err := server.Setup(); err != nil {
			return err
		}
		defer server.Stop()
		u.PrintInfo(fmt.Sprintf("HTTP server started on http://%s/", httpServerFlags.listenAddress))
		return server.Run()
	},
}

func init() {
	HTTPServerCmd.Flags().StringVarP(&httpServerFlags.listenAddress, "listen", "l", "0.0.0.0:8080", "Address and port to listen on")
	HTTPServerCmd.Flags().BoolVar(&httpServerFlags.enableUpload, "upload", false, "Enable file uploads via PUT requests")
}
