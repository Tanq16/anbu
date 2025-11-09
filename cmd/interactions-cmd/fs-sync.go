package interactionsCmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	fssync "github.com/tanq16/anbu/internal/interactions/fs-sync"
)

var (
	fsSyncClientFlags struct {
		addr   string
		dir    string
		ignore string
	}
	fsSyncServerFlags struct {
		port   int
		dir    string
		ignore string
	}
)

var FSSyncCmd = &cobra.Command{
	Use:   "fs-sync",
	Short: "Synchronize files between client and server",
	Long: `Synchronize files between a client and server using WebSocket.
The server maintains the source of truth, and clients sync to it.
Both client and server watch for file changes and propagate them in real-time.`,
}

var fsSyncClientCmd = &cobra.Command{
	Use:   "client",
	Short: "Run the fs-sync client",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		log.Info().Msgf("Starting fs-sync client: server_address=%s directory=%s ignores=%s", fsSyncClientFlags.addr, fsSyncClientFlags.dir, fsSyncClientFlags.ignore)
		cfg := fssync.ClientConfig{
			ServerAddr:  fsSyncClientFlags.addr,
			SyncDir:     fsSyncClientFlags.dir,
			IgnorePaths: fsSyncClientFlags.ignore,
		}
		c, err := fssync.NewClient(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize client")
		}
		c.Run(ctx)
	},
}

var fsSyncServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the fs-sync server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		log.Info().Msgf("Starting fs-sync server: port=%d directory=%s ignores=%s", fsSyncServerFlags.port, fsSyncServerFlags.dir, fsSyncServerFlags.ignore)
		cfg := fssync.ServerConfig{
			Port:        fsSyncServerFlags.port,
			SyncDir:     fsSyncServerFlags.dir,
			IgnorePaths: fsSyncServerFlags.ignore,
		}
		s, err := fssync.NewServer(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize server")
		}
		if err := s.Run(ctx); err != nil {
			log.Fatal().Err(err).Msg("Server exited with an error")
		}
	},
}

func init() {
	fsSyncClientCmd.Flags().StringVarP(&fsSyncClientFlags.addr, "addr", "a", "ws://localhost:8080/ws", "Address of the fs-sync server")
	fsSyncClientCmd.Flags().StringVarP(&fsSyncClientFlags.dir, "dir", "d", ".", "Directory to sync with the server")
	fsSyncClientCmd.Flags().StringVar(&fsSyncClientFlags.ignore, "ignore", "", "Comma-separated list of glob patterns to ignore for local changes (e.g., 'node_modules/*,*.log')")

	fsSyncServerCmd.Flags().IntVarP(&fsSyncServerFlags.port, "port", "p", 8080, "Port for the server to listen on")
	fsSyncServerCmd.Flags().StringVarP(&fsSyncServerFlags.dir, "dir", "d", ".", "Directory to sync (server's source of truth)")
	fsSyncServerCmd.Flags().StringVar(&fsSyncServerFlags.ignore, "ignore", "", "Comma-separated list of glob patterns to ignore (e.g., '.git/*,*.tmp')")

	FSSyncCmd.AddCommand(fsSyncClientCmd)
	FSSyncCmd.AddCommand(fsSyncServerCmd)
}
