package networkCmd

import (
	"github.com/rs/zerolog/log"
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
	Use:     "tunnel",
	Aliases: []string{},
	Short:   "Create TCP or SSH tunnels between local and remote endpoints",
	Long: `Provides tools for network tunneling.

Subcommands:
  tcp:   Create a simple TCP tunnel. Forwards traffic from a local port to a remote address.
  ssh:   Create an SSH forward tunnel. Forwards a local port to a remote address through an SSH server.
  rssh:  Create an SSH reverse tunnel. Forwards a remote port on an SSH server back to a local address.

Examples:
  # Forward local port 8080 to example.com:80
  anbu tunnel tcp -l localhost:8080 -r example.com:80

  # Forward local port 3306 to a database through an SSH jump host
  anbu tunnel ssh -l localhost:3306 -r db.internal:3306 -s jump.host:22 -u user -k ~/.ssh/id_rsa

  # Expose a local service (e.g., RDP on 3389) on a remote server's port 8001
  anbu tunnel rssh -l localhost:3389 -r 0.0.0.0:8001 -s remote.server:22 -u user -p "password"`,
}

var tcpTunnelCmd = &cobra.Command{
	Use:   "tcp",
	Short: "Create a TCP tunnel from a local port to a remote address",
	Run: func(cmd *cobra.Command, args []string) {
		if tunnelFlags.localAddr == "" {
			log.Fatal().Msg("Local address is required")
		}
		if tunnelFlags.remoteAddr == "" {
			log.Fatal().Msg("Remote address is required")
		}
		anbuNetwork.TCPTunnel(
			tunnelFlags.localAddr,
			tunnelFlags.remoteAddr,
			tunnelFlags.useTLS,
			tunnelFlags.insecureSkipVerify,
		)
	},
}

var sshTunnelCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Create an SSH forward tunnel through a jump host",
	Run: func(cmd *cobra.Command, args []string) {
		if tunnelFlags.remoteAddr == "" {
			log.Fatal().Msg("Remote address is required")
		}
		if tunnelFlags.sshAddr == "" {
			log.Fatal().Msg("SSH server address is required")
		}
		if tunnelFlags.sshUser == "" {
			log.Fatal().Msg("SSH username is required")
		}
		if tunnelFlags.sshPassword == "" && tunnelFlags.sshKeyPath == "" {
			log.Fatal().Msg("Either SSH password or key path is required")
		}
		var authMethods []ssh.AuthMethod
		if tunnelFlags.sshPassword != "" {
			authMethods = append(authMethods, anbuNetwork.TunnelSSHPassword(tunnelFlags.sshPassword))
		}
		if tunnelFlags.sshKeyPath != "" {
			keyAuth, err := anbuNetwork.TunnelSSHPrivateKey(tunnelFlags.sshKeyPath)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to load SSH key")
			}
			authMethods = append(authMethods, keyAuth)
		}
		anbuNetwork.SSHTunnel(
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
	Short: "Create a reverse SSH tunnel from a remote host to a local service",
	Run: func(cmd *cobra.Command, args []string) {
		if tunnelFlags.remoteAddr == "" {
			log.Fatal().Msg("Remote address is required")
		}
		if tunnelFlags.sshAddr == "" {
			log.Fatal().Msg("SSH server address is required")
		}
		if tunnelFlags.sshUser == "" {
			log.Fatal().Msg("SSH username is required")
		}
		if tunnelFlags.sshPassword == "" && tunnelFlags.sshKeyPath == "" {
			log.Fatal().Msg("Either SSH password or key path is required")
		}
		var authMethods []ssh.AuthMethod
		if tunnelFlags.sshPassword != "" {
			authMethods = append(authMethods, anbuNetwork.TunnelSSHPassword(tunnelFlags.sshPassword))
		}
		if tunnelFlags.sshKeyPath != "" {
			keyAuth, err := anbuNetwork.TunnelSSHPrivateKey(tunnelFlags.sshKeyPath)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to load SSH key")
			}
			authMethods = append(authMethods, keyAuth)
		}
		anbuNetwork.ReverseSSHTunnel(
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

	// TCP tunnel flags
	tcpTunnelCmd.Flags().StringVarP(&tunnelFlags.localAddr, "local", "l", "localhost:8000", "Local address:port to listen on")
	tcpTunnelCmd.Flags().StringVarP(&tunnelFlags.remoteAddr, "remote", "r", "", "Remote address to forward to")
	tcpTunnelCmd.Flags().BoolVar(&tunnelFlags.useTLS, "tls", false, "Use TLS for the remote connection")
	tcpTunnelCmd.Flags().BoolVar(&tunnelFlags.insecureSkipVerify, "insecure", false, "Skip TLS certificate verification")

	// SSH tunnel flags
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.localAddr, "local", "l", "localhost:8000", "Local address to listen on")
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.remoteAddr, "remote", "r", "", "Remote address to forward to")
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshAddr, "ssh", "s", "", "SSH server address (host:port)")
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshUser, "user", "u", "", "SSH username")
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshPassword, "password", "p", "", "SSH password")
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshKeyPath, "key", "k", "", "Path to SSH private key")

	// Reverse SSH tunnel flags
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.localAddr, "local", "l", "localhost:8000", "Local address to connect to")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.remoteAddr, "remote", "r", "", "Remote address to listen on")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshAddr, "ssh", "s", "", "SSH server address (host:port)")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshUser, "user", "u", "", "SSH username")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshPassword, "password", "p", "", "SSH password")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshKeyPath, "key", "k", "", "Path to SSH private key")
}
