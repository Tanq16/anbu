package networkCmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/wgproxy"
	u "github.com/tanq16/anbu/utils"
)

var wgProxyFlags struct {
	privateKey string
	peerKey    string
	endpoint   string
	address    string
	dns        string
	listen     string
	mtu        int
	keepalive  int
	timeout    time.Duration
}

var WgProxyCmd = &cobra.Command{
	Use:     "wg-proxy",
	Aliases: []string{"wgp"},
	Short:   "Start a userspace WireGuard SOCKS5 proxy",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		srv, err := wgproxy.New(wgproxy.Config{
			PrivateKey:    wgProxyFlags.privateKey,
			PeerPublicKey: wgProxyFlags.peerKey,
			Endpoint:      wgProxyFlags.endpoint,
			Address:       wgProxyFlags.address,
			DNS:           wgProxyFlags.dns,
			ListenAddr:    wgProxyFlags.listen,
			MTU:           wgProxyFlags.mtu,
			KeepAlive:     wgProxyFlags.keepalive,
			DialTimeout:   wgProxyFlags.timeout,
			Debug:         u.GlobalDebugFlag,
		})
		if err != nil {
			u.PrintFatal("failed to start WireGuard tunnel", err)
		}
		defer srv.Close()

		u.PrintInfo(fmt.Sprintf("WireGuard SOCKS5 proxy listening on %s", wgProxyFlags.listen))
		u.PrintInfo(fmt.Sprintf("Peer endpoint %s", wgProxyFlags.endpoint))
		if err := srv.Serve(ctx); err != nil {
			u.PrintFatal("proxy failed", err)
		}
		u.PrintInfo("proxy stopped")
		return nil
	},
}

func init() {
	WgProxyCmd.Flags().StringVarP(&wgProxyFlags.privateKey, "private-key", "k", "", "WireGuard private key (base64 or hex)")
	WgProxyCmd.Flags().StringVarP(&wgProxyFlags.peerKey, "peer-key", "p", "", "WireGuard peer public key (base64 or hex)")
	WgProxyCmd.Flags().StringVarP(&wgProxyFlags.endpoint, "endpoint", "e", "", "WireGuard peer endpoint (host:port)")
	WgProxyCmd.Flags().StringVarP(&wgProxyFlags.address, "address", "a", "", "Tunnel address assigned to this peer")
	WgProxyCmd.Flags().StringVarP(&wgProxyFlags.listen, "listen", "l", "127.0.0.1:1080", "Local proxy listen address")
	WgProxyCmd.Flags().StringVar(&wgProxyFlags.dns, "dns", "1.1.1.1", "DNS server used inside the tunnel")
	WgProxyCmd.Flags().IntVar(&wgProxyFlags.mtu, "mtu", 1420, "Tunnel MTU")
	WgProxyCmd.Flags().IntVar(&wgProxyFlags.keepalive, "keepalive", 25, "Persistent keepalive interval in seconds (0 disables)")
	WgProxyCmd.Flags().DurationVar(&wgProxyFlags.timeout, "timeout", 15*time.Second, "Dial timeout through the tunnel")
	WgProxyCmd.MarkFlagRequired("private-key")
	WgProxyCmd.MarkFlagRequired("peer-key")
	WgProxyCmd.MarkFlagRequired("endpoint")
	WgProxyCmd.MarkFlagRequired("address")
}
