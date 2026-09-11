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
	configFile   string
	privateKey   string
	peerKey      string
	presharedKey string
	endpoint     string
	address      string
	dns          string
	listen       string
	mtu          int
	keepalive    int
	timeout      time.Duration
}

var WgProxyCmd = &cobra.Command{
	Use:     "wg-proxy [config-file]",
	Aliases: []string{"wgp"},
	Short:   "Start a userspace WireGuard SOCKS5 proxy",
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		path := wgProxyFlags.configFile
		if len(args) == 1 {
			if path != "" {
				u.PrintFatal("give the config file as an argument or as --config-file, not both", nil)
			}
			path = args[0]
		}

		cfg, err := wgproxy.Load(path, wgproxy.Config{
			PrivateKey:    wgProxyFlags.privateKey,
			PeerPublicKey: wgProxyFlags.peerKey,
			PresharedKey:  wgProxyFlags.presharedKey,
			Endpoint:      wgProxyFlags.endpoint,
			Address:       wgProxyFlags.address,
			DNS:           wgProxyFlags.dns,
			ListenAddr:    wgProxyFlags.listen,
			MTU:           wgProxyFlags.mtu,
			KeepAlive:     wgProxyFlags.keepalive,
			DialTimeout:   wgProxyFlags.timeout,
			Debug:         u.GlobalDebugFlag,
		}, cmd.Flags().Changed)
		if err != nil {
			u.PrintFatal(err.Error(), err)
		}

		srv, err := wgproxy.New(cfg)
		if err != nil {
			u.PrintFatal(err.Error(), err)
		}
		defer srv.Close()

		u.PrintInfo(fmt.Sprintf("WireGuard SOCKS5 proxy listening on %s", cfg.ListenAddr))
		u.PrintInfo(fmt.Sprintf("Peer endpoint %s", cfg.Endpoint))
		if err := srv.Serve(ctx); err != nil {
			u.PrintFatal("proxy failed", err)
		}
		u.PrintInfo("proxy stopped")
		return nil
	},
}

func init() {
	WgProxyCmd.Flags().StringVar(&wgProxyFlags.configFile, "config-file", "", "WireGuard config file (wg-quick .conf)")
	WgProxyCmd.Flags().StringVarP(&wgProxyFlags.privateKey, "private-key", "k", "", "WireGuard private key (base64 or hex)")
	WgProxyCmd.Flags().StringVarP(&wgProxyFlags.peerKey, "peer-key", "p", "", "WireGuard peer public key (base64 or hex)")
	WgProxyCmd.Flags().StringVar(&wgProxyFlags.presharedKey, "preshared-key", "", "WireGuard preshared key (base64 or hex)")
	WgProxyCmd.Flags().StringVarP(&wgProxyFlags.endpoint, "endpoint", "e", "", "WireGuard peer endpoint (host:port)")
	WgProxyCmd.Flags().StringVarP(&wgProxyFlags.address, "address", "a", "", "Tunnel address assigned to this peer")
	WgProxyCmd.Flags().StringVarP(&wgProxyFlags.listen, "listen", "l", "127.0.0.1:8888", "Local SOCKS5 listen address")
	WgProxyCmd.Flags().StringVar(&wgProxyFlags.dns, "dns", "1.1.1.1", "DNS server used inside the tunnel")
	WgProxyCmd.Flags().IntVar(&wgProxyFlags.mtu, "mtu", 1420, "Tunnel MTU")
	WgProxyCmd.Flags().IntVar(&wgProxyFlags.keepalive, "keepalive", 25, "Persistent keepalive interval in seconds (0 disables)")
	WgProxyCmd.Flags().DurationVar(&wgProxyFlags.timeout, "timeout", 15*time.Second, "Dial timeout through the tunnel")
}
