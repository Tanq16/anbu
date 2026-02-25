package networkCmd

import (
	"github.com/spf13/cobra"
	anbuNetwork "github.com/tanq16/anbu/internal/network"
	u "github.com/tanq16/anbu/internal/utils"
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
}

var tcpTunnelCmd = &cobra.Command{
	Use:   "tcp",
	Short: "Create a TCP tunnel from a local port to a remote address with optional TLS support",
	Run: func(cmd *cobra.Command, args []string) {
		if tunnelFlags.localAddr == "" {
			u.PrintFatal("Local address is required", nil)
		}
		if tunnelFlags.remoteAddr == "" {
			u.PrintFatal("Remote address is required", nil)
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
	Short: "Create an SSH forward tunnel through a jump host with password or key-based authentication",
	Run: func(cmd *cobra.Command, args []string) {
		if tunnelFlags.remoteAddr == "" {
			u.PrintFatal("Remote address is required", nil)
		}
		if tunnelFlags.sshAddr == "" {
			u.PrintFatal("SSH server address is required", nil)
		}
		if tunnelFlags.sshUser == "" {
			u.PrintFatal("SSH username is required", nil)
		}
		if tunnelFlags.sshPassword == "" && tunnelFlags.sshKeyPath == "" {
			u.PrintFatal("Either SSH password or key path is required", nil)
		}
		var authMethods []ssh.AuthMethod
		if tunnelFlags.sshPassword != "" {
			authMethods = append(authMethods, anbuNetwork.TunnelSSHPassword(tunnelFlags.sshPassword))
		}
		if tunnelFlags.sshKeyPath != "" {
			keyAuth, err := anbuNetwork.TunnelSSHPrivateKey(tunnelFlags.sshKeyPath)
			if err != nil {
				u.PrintFatal("Failed to load SSH key", err)
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
	Short: "Create a reverse SSH tunnel from a remote host to a local service with password or key-based authentication",
	Run: func(cmd *cobra.Command, args []string) {
		if tunnelFlags.remoteAddr == "" {
			u.PrintFatal("Remote address is required", nil)
		}
		if tunnelFlags.sshAddr == "" {
			u.PrintFatal("SSH server address is required", nil)
		}
		if tunnelFlags.sshUser == "" {
			u.PrintFatal("SSH username is required", nil)
		}
		if tunnelFlags.sshPassword == "" && tunnelFlags.sshKeyPath == "" {
			u.PrintFatal("Either SSH password or key path is required", nil)
		}
		var authMethods []ssh.AuthMethod
		if tunnelFlags.sshPassword != "" {
			authMethods = append(authMethods, anbuNetwork.TunnelSSHPassword(tunnelFlags.sshPassword))
		}
		if tunnelFlags.sshKeyPath != "" {
			keyAuth, err := anbuNetwork.TunnelSSHPrivateKey(tunnelFlags.sshKeyPath)
			if err != nil {
				u.PrintFatal("Failed to load SSH key", err)
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
