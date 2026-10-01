package cmd

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/server"
)

var serveFlags struct {
	host    string
	port    int
	dataDir string
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the anbu server",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		srv, err := server.New(server.Config{Host: serveFlags.host, Port: serveFlags.port, DataDir: serveFlags.dataDir})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open data directory")
		}
		if err := srv.Setup(); err != nil {
			log.Fatal().Err(err).Msg("failed to set up server")
		}
		if err := srv.Run(); err != nil {
			log.Fatal().Err(err).Msg("server error")
		}
	},
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "anbu", "data")
}

func init() {
	serveCmd.Flags().StringVarP(&serveFlags.host, "host", "H", "0.0.0.0", "Host to bind to")
	serveCmd.Flags().IntVarP(&serveFlags.port, "port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringVarP(&serveFlags.dataDir, "data-dir", "d", defaultDataDir(), "Data directory")
}
