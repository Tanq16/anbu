package interactionsCmd

import (
	"fmt"

	"github.com/spf13/cobra"
	fssync "github.com/tanq16/anbu/internal/interactions/fs-sync"
	u "github.com/tanq16/anbu/utils"
)

var (
	fsSyncSendFlags struct {
		listen    bool
		connect   string
		port      int
		dir       string
		ignore    string
		enableTLS bool
		insecure  bool
	}
	fsSyncReceiveFlags struct {
		listen    bool
		connect   string
		port      int
		dir       string
		ignore    string
		delete    bool
		dryRun    bool
		enableTLS bool
		insecure  bool
	}
)

var FSSyncCmd = &cobra.Command{
	Use:   "fs-sync",
	Short: "One-shot file synchronization over HTTP/HTTPS with optional TLS, ignore patterns, and dry-run mode",
}

var fsSyncSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send files to a peer (use --listen or --connect)",
	Run: func(cmd *cobra.Command, args []string) {
		if fsSyncSendFlags.listen && fsSyncSendFlags.connect != "" {
			u.PrintFatal("cannot specify both --listen and --connect", nil)
		}
		if !fsSyncSendFlags.listen && fsSyncSendFlags.connect == "" {
			u.PrintFatal("must specify either --listen or --connect", nil)
		}
		if fsSyncSendFlags.listen {
			protocol := "http"
			if fsSyncSendFlags.enableTLS {
				protocol = "https"
			}
			u.PrintInfo(fmt.Sprintf("Starting fs-sync server (send): %s://localhost:%d directory=%s", protocol, fsSyncSendFlags.port, fsSyncSendFlags.dir))
			cfg := fssync.ServerConfig{
				Port:        fsSyncSendFlags.port,
				SyncDir:     fsSyncSendFlags.dir,
				IgnorePaths: fsSyncSendFlags.ignore,
				EnableTLS:   fsSyncSendFlags.enableTLS,
				Mode:        "send",
			}
			s, err := fssync.NewServer(cfg)
			if err != nil {
				u.PrintFatal("Failed to initialize server", err)
			}
			if err := s.Run(); err != nil {
				u.PrintFatal("Failed to run server", err)
			}
		} else {
			cfg := fssync.ClientConfig{
				ServerAddr:  fsSyncSendFlags.connect,
				SyncDir:     fsSyncSendFlags.dir,
				Insecure:    fsSyncSendFlags.insecure,
				Mode:        "send",
				IgnorePaths: fsSyncSendFlags.ignore,
			}
			c, err := fssync.NewClient(cfg)
			if err != nil {
				u.PrintFatal("Failed to initialize client", err)
			}
			if err := c.Run(); err != nil {
				u.PrintFatal("Failed to send", err)
			}
		}
	},
}

var fsSyncReceiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "Receive files from a peer (use --listen or --connect)",
	Run: func(cmd *cobra.Command, args []string) {
		if fsSyncReceiveFlags.listen && fsSyncReceiveFlags.connect != "" {
			u.PrintFatal("cannot specify both --listen and --connect", nil)
		}
		if !fsSyncReceiveFlags.listen && fsSyncReceiveFlags.connect == "" {
			u.PrintFatal("must specify either --listen or --connect", nil)
		}
		if fsSyncReceiveFlags.listen {
			protocol := "http"
			if fsSyncReceiveFlags.enableTLS {
				protocol = "https"
			}
			u.PrintInfo(fmt.Sprintf("Starting fs-sync server (receive): %s://localhost:%d directory=%s", protocol, fsSyncReceiveFlags.port, fsSyncReceiveFlags.dir))
			cfg := fssync.ServerConfig{
				Port:        fsSyncReceiveFlags.port,
				SyncDir:     fsSyncReceiveFlags.dir,
				IgnorePaths: fsSyncReceiveFlags.ignore,
				EnableTLS:   fsSyncReceiveFlags.enableTLS,
				Mode:        "receive",
				DeleteExtra: fsSyncReceiveFlags.delete,
				DryRun:      fsSyncReceiveFlags.dryRun,
			}
			s, err := fssync.NewServer(cfg)
			if err != nil {
				u.PrintFatal("Failed to initialize server", err)
			}
			if err := s.Run(); err != nil {
				u.PrintFatal("Failed to run server", err)
			}
		} else {
			cfg := fssync.ClientConfig{
				ServerAddr:  fsSyncReceiveFlags.connect,
				SyncDir:     fsSyncReceiveFlags.dir,
				DeleteExtra: fsSyncReceiveFlags.delete,
				Insecure:    fsSyncReceiveFlags.insecure,
				DryRun:      fsSyncReceiveFlags.dryRun,
				Mode:        "receive",
				IgnorePaths: fsSyncReceiveFlags.ignore,
			}
			c, err := fssync.NewClient(cfg)
			if err != nil {
				u.PrintFatal("Failed to initialize client", err)
			}
			if err := c.Run(); err != nil {
				u.PrintFatal("Failed to receive", err)
			}
		}
	},
}

func init() {
	fsSyncSendCmd.Flags().BoolVarP(&fsSyncSendFlags.listen, "listen", "l", false, "Listen for incoming connections")
	fsSyncSendCmd.Flags().StringVarP(&fsSyncSendFlags.connect, "connect", "c", "", "Connect to a listening receiver (URL)")
	fsSyncSendCmd.Flags().IntVarP(&fsSyncSendFlags.port, "port", "p", 8080, "Port to listen on (with --listen)")
	fsSyncSendCmd.Flags().StringVarP(&fsSyncSendFlags.dir, "dir", "d", ".", "Directory to send")
	fsSyncSendCmd.Flags().StringVar(&fsSyncSendFlags.ignore, "ignore", "", "Comma-separated patterns to ignore (e.g., '.git,node_modules')")
	fsSyncSendCmd.Flags().BoolVarP(&fsSyncSendFlags.enableTLS, "tls", "t", false, "Enable HTTPS with self-signed cert (with --listen)")
	fsSyncSendCmd.Flags().BoolVarP(&fsSyncSendFlags.insecure, "insecure", "k", false, "Skip TLS verification (with --connect)")

	fsSyncReceiveCmd.Flags().BoolVarP(&fsSyncReceiveFlags.listen, "listen", "l", false, "Listen for incoming connections")
	fsSyncReceiveCmd.Flags().StringVarP(&fsSyncReceiveFlags.connect, "connect", "c", "", "Connect to a listening sender (URL)")
	fsSyncReceiveCmd.Flags().IntVarP(&fsSyncReceiveFlags.port, "port", "p", 8080, "Port to listen on (with --listen)")
	fsSyncReceiveCmd.Flags().StringVarP(&fsSyncReceiveFlags.dir, "dir", "d", ".", "Directory to receive into")
	fsSyncReceiveCmd.Flags().StringVar(&fsSyncReceiveFlags.ignore, "ignore", "", "Comma-separated patterns to ignore (e.g., '.git,node_modules')")
	fsSyncReceiveCmd.Flags().BoolVar(&fsSyncReceiveFlags.delete, "delete", false, "Delete local files not present on sender")
	fsSyncReceiveCmd.Flags().BoolVarP(&fsSyncReceiveFlags.dryRun, "dry-run", "r", false, "Show what would be synced without doing it")
	fsSyncReceiveCmd.Flags().BoolVarP(&fsSyncReceiveFlags.enableTLS, "tls", "t", false, "Enable HTTPS with self-signed cert (with --listen)")
	fsSyncReceiveCmd.Flags().BoolVarP(&fsSyncReceiveFlags.insecure, "insecure", "k", false, "Skip TLS verification (with --connect)")

	FSSyncCmd.AddCommand(fsSyncSendCmd)
	FSSyncCmd.AddCommand(fsSyncReceiveCmd)
}
