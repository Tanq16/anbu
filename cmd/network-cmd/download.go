package networkCmd

import (
	"context"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/download"
	u "github.com/tanq16/anbu/utils"
)

var downloadFlags struct {
	output      string
	proxy       string
	userAgent   string
	headers     []string
	connections int
}

var DownloadCmd = &cobra.Command{
	Use:     "download <url>",
	Aliases: []string{"dl"},
	Short:   "Download a file over HTTP",
	Args:    cobra.ExactArgs(1),
	Run:     runDownload,
}

func runDownload(cmd *cobra.Command, args []string) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()

	cfg := download.Config{
		URL:         args[0],
		OutputPath:  downloadFlags.output,
		Connections: downloadFlags.connections,
		ProxyURL:    downloadFlags.proxy,
		UserAgent:   downloadFlags.userAgent,
		Headers:     download.ParseHeaders(downloadFlags.headers),
	}
	plan, client, err := download.Prepare(ctx, cfg)
	if download.AlreadyComplete(err) {
		u.PrintSuccess(download.AlreadyCompletePath(err) + " already exists")
		return
	}
	if err != nil {
		u.PrintFatal("failed to start download", err)
	}

	m := u.NewMeter("Downloading", filepath.Base(plan.OutputPath), plan.Size, u.UnitBytes)
	if err := plan.Execute(ctx, client, m); err != nil {
		m.Fail(err)
		u.PrintFatal("download failed", err)
	}
	m.Done()
}

func init() {
	DownloadCmd.Flags().StringVarP(&downloadFlags.output, "output", "o", "", "Output file path")
	DownloadCmd.Flags().StringVarP(&downloadFlags.proxy, "proxy", "p", "", "HTTP/HTTPS proxy URL")
	DownloadCmd.Flags().StringVarP(&downloadFlags.userAgent, "user-agent", "a", "Anbu-CLI", "User agent")
	DownloadCmd.Flags().StringArrayVarP(&downloadFlags.headers, "header", "H", nil, "Custom header as Key: Value (repeatable)")
	DownloadCmd.Flags().IntVarP(&downloadFlags.connections, "connections", "c", 8, "Parallel connections for a range-capable download")
}
