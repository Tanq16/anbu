package interactionsCmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	fsSyncServeFlags struct {
		port   int
		dir    string
		ignore string
	}
	fsSyncSyncFlags struct {
		server string
		dir    string
		ignore string
		delete bool
		dryRun bool
	}
)

var FSSyncCmd = &cobra.Command{
	Use:   "fs-sync",
	Short: "One-shot file synchronization over WebSocket",
	Long: `One-shot file synchronization between two machines.
Server side runs 'serve' and waits for one client connection.
Client side runs 'sync' to connect and sync files.
Both commands exit after sync completes.`,
}

var fsSyncServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve files and wait for one sync client",
	Long: `Start a server that waits for one client connection, serves files, and exits.
Run this on the machine with the source files.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("fs-sync serve: not implemented yet")
	},
}

var fsSyncSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Connect to server and sync files",
	Long: `Connect to a server, receive file manifest, sync files, and exit.
Run this on the machine that wants to receive the files.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("fs-sync sync: not implemented yet")
	},
}

func init() {
	fsSyncServeCmd.Flags().IntVarP(&fsSyncServeFlags.port, "port", "p", 8080, "Port to listen on")
	fsSyncServeCmd.Flags().StringVarP(&fsSyncServeFlags.dir, "dir", "d", ".", "Directory to serve")
	fsSyncServeCmd.Flags().StringVar(&fsSyncServeFlags.ignore, "ignore", "", "Comma-separated patterns to ignore (e.g., '.git,node_modules')")

	fsSyncSyncCmd.Flags().StringVarP(&fsSyncSyncFlags.server, "server", "s", "ws://localhost:8080/ws", "Server address")
	fsSyncSyncCmd.Flags().StringVarP(&fsSyncSyncFlags.dir, "dir", "d", ".", "Local directory to sync to")
	fsSyncSyncCmd.Flags().StringVar(&fsSyncSyncFlags.ignore, "ignore", "", "Comma-separated patterns to ignore")
	fsSyncSyncCmd.Flags().BoolVar(&fsSyncSyncFlags.delete, "delete", false, "Delete local files not present on server")
	fsSyncSyncCmd.Flags().BoolVar(&fsSyncSyncFlags.dryRun, "dry-run", false, "Show what would be synced without doing it")

	FSSyncCmd.AddCommand(fsSyncServeCmd)
	FSSyncCmd.AddCommand(fsSyncSyncCmd)
}
