package networkCmd

import (
	"github.com/spf13/cobra"
	anbuNetwork "github.com/tanq16/anbu/internal/network"
	"github.com/tanq16/anbu/utils"
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
	Short: "Create tunnels between local and remote endpoints",
}

var tcpTunnelCmd = &cobra.Command{
	Use:   "tcp",
	Short: "Create a TCP tunnel from local to remote",
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("tunnel-tcp")
		if tunnelFlags.localAddr == "" {
			logger.Fatal().Msg("Local address is required")
		}
		if tunnelFlags.remoteAddr == "" {
			logger.Fatal().Msg("Remote address is required")
		}
		err := anbuNetwork.TCPTunnel(
			tunnelFlags.localAddr,
			tunnelFlags.remoteAddr,
			tunnelFlags.useTLS,
			tunnelFlags.insecureSkipVerify,
		)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create TCP tunnel")
		}
	},
}

var sshTunnelCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Create an SSH tunnel from local to remote",
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("tunnel-ssh")
		if tunnelFlags.remoteAddr == "" {
			logger.Fatal().Msg("Remote address is required")
		}
		if tunnelFlags.sshAddr == "" {
			logger.Fatal().Msg("SSH server address is required")
		}
		if tunnelFlags.sshUser == "" {
			logger.Fatal().Msg("SSH username is required")
		}
		if tunnelFlags.sshPassword == "" && tunnelFlags.sshKeyPath == "" {
			logger.Fatal().Msg("Either SSH password or key path is required")
		}
		var authMethods []ssh.AuthMethod
		if tunnelFlags.sshPassword != "" {
			authMethods = append(authMethods, anbuNetwork.TunnelSSHPassword(tunnelFlags.sshPassword))
			logger.Debug().Msg("Using SSH password authentication")
		}
		if tunnelFlags.sshKeyPath != "" {
			keyAuth, err := anbuNetwork.TunnelSSHPrivateKey(tunnelFlags.sshKeyPath)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to load SSH key")
			}
			authMethods = append(authMethods, keyAuth)
			logger.Debug().Msg("Using SSH key authentication")
		}
		err := anbuNetwork.SSHTunnel(
			tunnelFlags.localAddr,
			tunnelFlags.remoteAddr,
			tunnelFlags.sshAddr,
			tunnelFlags.sshUser,
			authMethods,
		)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create SSH tunnel")
		}
	},
}

var reverseTcpTunnelCmd = &cobra.Command{
	Use:   "rtcp",
	Short: "Create a reverse TCP tunnel from remote to local",
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("tunnel-rtcp")
		if tunnelFlags.localAddr == "" {
			logger.Fatal().Msg("Local address is required")
		}
		if tunnelFlags.remoteAddr == "" {
			logger.Fatal().Msg("Remote address is required")
		}
		err := anbuNetwork.ReverseTCPTunnel(
			tunnelFlags.localAddr,
			tunnelFlags.remoteAddr,
			tunnelFlags.useTLS,
			tunnelFlags.insecureSkipVerify,
		)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create reverse TCP tunnel")
		}
	},
}

var reverseSshTunnelCmd = &cobra.Command{
	Use:   "rssh",
	Short: "Create a reverse SSH tunnel from remote to local",
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("tunnel-rssh")
		if tunnelFlags.remoteAddr == "" {
			logger.Fatal().Msg("Remote address is required")
		}
		if tunnelFlags.sshAddr == "" {
			logger.Fatal().Msg("SSH server address is required")
		}
		if tunnelFlags.sshUser == "" {
			logger.Fatal().Msg("SSH username is required")
		}
		if tunnelFlags.sshPassword == "" && tunnelFlags.sshKeyPath == "" {
			logger.Fatal().Msg("Either SSH password or key path is required")
		}
		var authMethods []ssh.AuthMethod
		if tunnelFlags.sshPassword != "" {
			authMethods = append(authMethods, anbuNetwork.TunnelSSHPassword(tunnelFlags.sshPassword))
			logger.Debug().Msg("Using SSH password authentication")
		}
		if tunnelFlags.sshKeyPath != "" {
			keyAuth, err := anbuNetwork.TunnelSSHPrivateKey(tunnelFlags.sshKeyPath)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to load SSH key")
			}
			authMethods = append(authMethods, keyAuth)
			logger.Debug().Msg("Using SSH key authentication")
		}
		err := anbuNetwork.ReverseSSHTunnel(
			tunnelFlags.localAddr,
			tunnelFlags.remoteAddr,
			tunnelFlags.sshAddr,
			tunnelFlags.sshUser,
			authMethods,
		)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create reverse SSH tunnel")
		}
	},
}

func init() {
	TunnelCmd.AddCommand(tcpTunnelCmd)
	TunnelCmd.AddCommand(sshTunnelCmd)
	TunnelCmd.AddCommand(reverseTcpTunnelCmd)
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
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshKeyPath, "key", "k", "", "SSH private key path")

	// Reverse TCP tunnel flags
	reverseTcpTunnelCmd.Flags().StringVarP(&tunnelFlags.localAddr, "local", "l", "localhost:8000", "Local address:port to connect to")
	reverseTcpTunnelCmd.Flags().StringVarP(&tunnelFlags.remoteAddr, "remote", "r", "", "Remote address:port to listen on")
	reverseTcpTunnelCmd.Flags().BoolVar(&tunnelFlags.useTLS, "tls", false, "Use TLS for the connection")
	reverseTcpTunnelCmd.Flags().BoolVar(&tunnelFlags.insecureSkipVerify, "insecure", false, "Skip TLS certificate verification")

	// Reverse SSH tunnel flags
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.localAddr, "local", "l", "localhost:8000", "Local address to connect to")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.remoteAddr, "remote", "r", "", "Remote address to listen on")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshAddr, "ssh", "s", "", "SSH server address (host:port)")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshUser, "user", "u", "", "SSH username")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshPassword, "password", "p", "", "SSH password")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshKeyPath, "key", "k", "", "SSH private key path")
}
