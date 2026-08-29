package networkCmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	anbuNetwork "github.com/tanq16/anbu/internal/network"
	"golang.org/x/crypto/ssh"
)

var tunnelFlags struct {
	localAddr          string
	remoteAddr         string
	useTLS             bool
	insecureSkipVerify bool
	sshAddr            string
	sshUser            string
	sshPassword        string
	sshKeyPath         string
}

var TunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "Create TCP or SSH tunnels between local and remote endpoints",
}

var tcpTunnelCmd = &cobra.Command{
	Use:   "tcp",
	Short: "Create a TCP tunnel from a local port to a remote address with optional TLS support",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		return anbuNetwork.TCPTunnel(
			ctx,
			tunnelFlags.localAddr,
			tunnelFlags.remoteAddr,
			tunnelFlags.useTLS,
			tunnelFlags.insecureSkipVerify,
		)
	},
}

var sshTunnelCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Create an SSH forward tunnel through a jump host with password or key-based authentication",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var authMethods []ssh.AuthMethod
		if tunnelFlags.sshPassword != "" {
			authMethods = append(authMethods, anbuNetwork.TunnelSSHPassword(tunnelFlags.sshPassword))
		}
		if tunnelFlags.sshKeyPath != "" {
			keyAuth, err := anbuNetwork.TunnelSSHPrivateKey(tunnelFlags.sshKeyPath)
			if err != nil {
				return err
			}
			authMethods = append(authMethods, keyAuth)
		}
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		return anbuNetwork.SSHTunnel(
			ctx,
			tunnelFlags.localAddr,
			tunnelFlags.remoteAddr,
			tunnelFlags.sshAddr,
			tunnelFlags.sshUser,
			authMethods,
		)
	},
}

var reverseSshTunnelCmd = &cobra.Command{
	Use:   "rssh",
	Short: "Create a reverse SSH tunnel from a remote host to a local service with password or key-based authentication",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var authMethods []ssh.AuthMethod
		if tunnelFlags.sshPassword != "" {
			authMethods = append(authMethods, anbuNetwork.TunnelSSHPassword(tunnelFlags.sshPassword))
		}
		if tunnelFlags.sshKeyPath != "" {
			keyAuth, err := anbuNetwork.TunnelSSHPrivateKey(tunnelFlags.sshKeyPath)
			if err != nil {
				return err
			}
			authMethods = append(authMethods, keyAuth)
		}
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		return anbuNetwork.ReverseSSHTunnel(
			ctx,
			tunnelFlags.localAddr,
			tunnelFlags.remoteAddr,
			tunnelFlags.sshAddr,
			tunnelFlags.sshUser,
			authMethods,
		)
	},
}

func init() {
	TunnelCmd.AddCommand(tcpTunnelCmd)
	TunnelCmd.AddCommand(sshTunnelCmd)
	TunnelCmd.AddCommand(reverseSshTunnelCmd)

	tcpTunnelCmd.Flags().StringVarP(&tunnelFlags.localAddr, "local", "l", "localhost:8000", "Local address:port to listen on")
	tcpTunnelCmd.Flags().StringVarP(&tunnelFlags.remoteAddr, "remote", "r", "", "Remote address to forward to")
	tcpTunnelCmd.Flags().BoolVar(&tunnelFlags.useTLS, "tls", false, "Use TLS for the remote connection")
	tcpTunnelCmd.Flags().BoolVar(&tunnelFlags.insecureSkipVerify, "insecure", false, "Skip TLS certificate verification")
	_ = tcpTunnelCmd.MarkFlagRequired("remote")

	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.localAddr, "local", "l", "localhost:8000", "Local address to listen on")
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.remoteAddr, "remote", "r", "", "Remote address to forward to")
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshAddr, "ssh", "s", "", "SSH server address (host:port)")
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshUser, "user", "u", "", "SSH username")
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshPassword, "password", "p", "", "SSH password")
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshKeyPath, "key", "k", "", "Path to SSH private key")
	_ = sshTunnelCmd.MarkFlagRequired("remote")
	_ = sshTunnelCmd.MarkFlagRequired("ssh")
	_ = sshTunnelCmd.MarkFlagRequired("user")
	sshTunnelCmd.MarkFlagsOneRequired("password", "key")

	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.localAddr, "local", "l", "localhost:8000", "Local address to connect to")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.remoteAddr, "remote", "r", "", "Remote address to listen on")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshAddr, "ssh", "s", "", "SSH server address (host:port)")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshUser, "user", "u", "", "SSH username")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshPassword, "password", "p", "", "SSH password")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshKeyPath, "key", "k", "", "Path to SSH private key")
	_ = reverseSshTunnelCmd.MarkFlagRequired("remote")
	_ = reverseSshTunnelCmd.MarkFlagRequired("ssh")
	_ = reverseSshTunnelCmd.MarkFlagRequired("user")
	reverseSshTunnelCmd.MarkFlagsOneRequired("password", "key")
}
