package interactionsCmd

import (
	"fmt"

	"github.com/spf13/cobra"
	fssync "github.com/tanq16/anbu/internal/interactions/fs-sync"
	u "github.com/tanq16/anbu/internal/utils"
)

var fsSyncServeFlags struct {
	mode      string
	port      int
	dir       string
	ignore    string
	enableTLS bool
	delete    bool
	dryRun    bool
}

var fsSyncClientFlags struct {
	dir      string
	ignore   string
	insecure bool
	delete   bool
	dryRun   bool
}

var FSSyncCmd = &cobra.Command{
	Use:   "fs-sync",
	Short: "One-shot bidirectional file synchronization over HTTP/HTTPS",
}

var fsSyncServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start an HTTP server for file sync (use --mode to set direction)",
	Run: func(cmd *cobra.Command, args []string) {
		if fsSyncServeFlags.mode != "send" && fsSyncServeFlags.mode != "receive" {
			u.PrintFatal("--mode must be 'send' or 'receive'", nil)
		}
		protocol := "http"
		if fsSyncServeFlags.enableTLS {
			protocol = "https"
		}
		u.PrintInfo(fmt.Sprintf("Starting fs-sync server (mode: %s): %s://localhost:%d directory=%s", fsSyncServeFlags.mode, protocol, fsSyncServeFlags.port, fsSyncServeFlags.dir))
		cfg := fssync.ServerConfig{
			Port:        fsSyncServeFlags.port,
			SyncDir:     fsSyncServeFlags.dir,
			IgnorePaths: fsSyncServeFlags.ignore,
			EnableTLS:   fsSyncServeFlags.enableTLS,
			Mode:        fsSyncServeFlags.mode,
			DeleteExtra: fsSyncServeFlags.delete,
			DryRun:      fsSyncServeFlags.dryRun,
		}
		s, err := fssync.NewServer(cfg)
		if err != nil {
			u.PrintFatal("Failed to initialize server", err)
		}
		if err := s.Run(); err != nil {
			u.PrintFatal("Failed to run server", err)
		}
	},
}

var fsSyncClientCmd = &cobra.Command{
	Use:   "client <server-url>",
	Short: "Connect to an fs-sync server and sync files (direction auto-detected)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := fssync.ClientConfig{
			ServerAddr:  args[0],
			SyncDir:     fsSyncClientFlags.dir,
			DeleteExtra: fsSyncClientFlags.delete,
			Insecure:    fsSyncClientFlags.insecure,
			DryRun:      fsSyncClientFlags.dryRun,
			IgnorePaths: fsSyncClientFlags.ignore,
		}
		c, err := fssync.NewClient(cfg)
		if err != nil {
			u.PrintFatal("Failed to initialize client", err)
		}
		if err := c.Run(); err != nil {
			u.PrintFatal("Sync failed", err)
		}
	},
}

func init() {
	fsSyncServeCmd.Flags().StringVarP(&fsSyncServeFlags.mode, "mode", "m", "send", "Sync mode: 'send' (serve files) or 'receive' (accept files)")
	fsSyncServeCmd.Flags().IntVarP(&fsSyncServeFlags.port, "port", "p", 8080, "Port to listen on")
	fsSyncServeCmd.Flags().StringVarP(&fsSyncServeFlags.dir, "dir", "d", ".", "Directory to sync")
	fsSyncServeCmd.Flags().StringVar(&fsSyncServeFlags.ignore, "ignore", "", "Comma-separated patterns to ignore (e.g., '.git,node_modules')")
	fsSyncServeCmd.Flags().BoolVarP(&fsSyncServeFlags.enableTLS, "tls", "t", false, "Enable HTTPS with self-signed cert")
	fsSyncServeCmd.Flags().BoolVar(&fsSyncServeFlags.delete, "delete", false, "Delete extra files not present on sender (receive mode only)")
	fsSyncServeCmd.Flags().BoolVarP(&fsSyncServeFlags.dryRun, "dry-run", "r", false, "Show what would be synced without doing it (receive mode only)")

	fsSyncClientCmd.Flags().StringVarP(&fsSyncClientFlags.dir, "dir", "d", ".", "Local directory to sync")
	fsSyncClientCmd.Flags().StringVar(&fsSyncClientFlags.ignore, "ignore", "", "Comma-separated patterns to ignore (e.g., '.git,node_modules')")
	fsSyncClientCmd.Flags().BoolVarP(&fsSyncClientFlags.insecure, "insecure", "k", false, "Skip TLS certificate verification")
	fsSyncClientCmd.Flags().BoolVar(&fsSyncClientFlags.delete, "delete", false, "Delete local files not present on server (when pulling)")
	fsSyncClientCmd.Flags().BoolVarP(&fsSyncClientFlags.dryRun, "dry-run", "r", false, "Show what would be synced without doing it (when pulling)")

	FSSyncCmd.AddCommand(fsSyncServeCmd)
	FSSyncCmd.AddCommand(fsSyncClientCmd)
}
