package interactionsCmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	fssync "github.com/tanq16/anbu/internal/interactions/fs-sync"
)

var (
	fsSyncServeFlags struct {
		port      int
		dir       string
		ignore    string
		enableTLS bool
	}
	fsSyncSyncFlags struct {
		server   string
		dir      string
		delete   bool
		dryRun   bool
		insecure bool
	}
)

var FSSyncCmd = &cobra.Command{
	Use:   "fs-sync",
	Short: "One-shot file synchronization over HTTP/HTTPS with optional TLS, ignore patterns, and dry-run mode",
}

var fsSyncServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve files to sync to a client",
	Run: func(cmd *cobra.Command, args []string) {
		protocol := "http"
		if fsSyncServeFlags.enableTLS {
			protocol = "https"
		}
		log.Info().Msgf("Starting fs-sync server: %s://localhost:%d directory=%s",
			protocol, fsSyncServeFlags.port, fsSyncServeFlags.dir)
		cfg := fssync.ServerConfig{
			Port:        fsSyncServeFlags.port,
			SyncDir:     fsSyncServeFlags.dir,
			IgnorePaths: fsSyncServeFlags.ignore,
			EnableTLS:   fsSyncServeFlags.enableTLS,
		}
		s, err := fssync.NewServer(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize server")
		}
		if err := s.Run(); err != nil {
			log.Fatal().Err(err).Msg("Server error")
		}
	},
}

var fsSyncSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Connect to a server and sync files",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := fssync.ClientConfig{
			ServerAddr:  fsSyncSyncFlags.server,
			SyncDir:     fsSyncSyncFlags.dir,
			DeleteExtra: fsSyncSyncFlags.delete,
			Insecure:    fsSyncSyncFlags.insecure,
			DryRun:      fsSyncSyncFlags.dryRun,
		}
		c, err := fssync.NewClient(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize client")
		}
		if err := c.Run(); err != nil {
			log.Fatal().Err(err).Msg("Sync failed")
		}
	},
}

func init() {
	fsSyncServeCmd.Flags().IntVarP(&fsSyncServeFlags.port, "port", "p", 8080, "Port to listen on")
	fsSyncServeCmd.Flags().StringVarP(&fsSyncServeFlags.dir, "dir", "d", ".", "Directory to serve")
	fsSyncServeCmd.Flags().StringVar(&fsSyncServeFlags.ignore, "ignore", "", "Comma-separated patterns to ignore (e.g., '.git,node_modules')")
	fsSyncServeCmd.Flags().BoolVarP(&fsSyncServeFlags.enableTLS, "tls", "t", false, "Enable HTTPS with a self-signed certificate")

	fsSyncSyncCmd.Flags().StringVarP(&fsSyncSyncFlags.server, "server", "s", "http://localhost:8080", "Server URL (http:// or https://)")
	fsSyncSyncCmd.Flags().StringVarP(&fsSyncSyncFlags.dir, "dir", "d", ".", "Local directory to sync to")
	fsSyncSyncCmd.Flags().BoolVar(&fsSyncSyncFlags.delete, "delete", false, "Delete local files not present on server")
	fsSyncSyncCmd.Flags().BoolVarP(&fsSyncSyncFlags.dryRun, "dry-run", "r", false, "Show what would be synced without doing it")
	fsSyncSyncCmd.Flags().BoolVarP(&fsSyncSyncFlags.insecure, "insecure", "k", false, "Skip TLS certificate verification")

	FSSyncCmd.AddCommand(fsSyncServeCmd)
	FSSyncCmd.AddCommand(fsSyncSyncCmd)
}
