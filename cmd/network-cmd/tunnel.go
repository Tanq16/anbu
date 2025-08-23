package networkCmd

import (
	"os"

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
	Use:     "tunnel",
	Aliases: []string{},
	Short:   "Create tunnels between local and remote endpoints",
	Long: `Examples:
s
`,
}

var tcpTunnelCmd = &cobra.Command{
	Use:   "tcp",
	Short: "Create a TCP tunnel from local to remote",
	Run: func(cmd *cobra.Command, args []string) {
		if tunnelFlags.localAddr == "" {
			utils.PrintError("Local address is required")
			os.Exit(1)
		}
		if tunnelFlags.remoteAddr == "" {
			utils.PrintError("Remote address is required")
			os.Exit(1)
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
	Short: "Create an SSH tunnel from local to remote",
	Run: func(cmd *cobra.Command, args []string) {
		if tunnelFlags.remoteAddr == "" {
			utils.PrintError("Remote address is required")
			os.Exit(1)
		}
		if tunnelFlags.sshAddr == "" {
			utils.PrintError("SSH server address is required")
			os.Exit(1)
		}
		if tunnelFlags.sshUser == "" {
			utils.PrintError("SSH username is required")
			os.Exit(1)
		}
		if tunnelFlags.sshPassword == "" && tunnelFlags.sshKeyPath == "" {
			utils.PrintError("Either SSH password or key path is required")
			os.Exit(1)
		}
		var authMethods []ssh.AuthMethod
		if tunnelFlags.sshPassword != "" {
			authMethods = append(authMethods, anbuNetwork.TunnelSSHPassword(tunnelFlags.sshPassword))
		}
		if tunnelFlags.sshKeyPath != "" {
			keyAuth, err := anbuNetwork.TunnelSSHPrivateKey(tunnelFlags.sshKeyPath)
			if err != nil {
				utils.PrintError("Failed to load SSH key: " + err.Error())
				os.Exit(1)
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
	Short: "Create a reverse SSH tunnel from remote to local",
	Run: func(cmd *cobra.Command, args []string) {
		if tunnelFlags.remoteAddr == "" {
			utils.PrintError("Remote address is required")
			os.Exit(1)
		}
		if tunnelFlags.sshAddr == "" {
			utils.PrintError("SSH server address is required")
			os.Exit(1)
		}
		if tunnelFlags.sshUser == "" {
			utils.PrintError("SSH username is required")
			os.Exit(1)
		}
		if tunnelFlags.sshPassword == "" && tunnelFlags.sshKeyPath == "" {
			utils.PrintError("Either SSH password or key path is required")
			os.Exit(1)
		}
		var authMethods []ssh.AuthMethod
		if tunnelFlags.sshPassword != "" {
			authMethods = append(authMethods, anbuNetwork.TunnelSSHPassword(tunnelFlags.sshPassword))
		}
		if tunnelFlags.sshKeyPath != "" {
			keyAuth, err := anbuNetwork.TunnelSSHPrivateKey(tunnelFlags.sshKeyPath)
			if err != nil {
				utils.PrintError("Failed to load SSH key: " + err.Error())
				os.Exit(1)
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
	sshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshKeyPath, "key", "k", "", "SSH private key path")

	// Reverse SSH tunnel flags
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.localAddr, "local", "l", "localhost:8000", "Local address to connect to")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.remoteAddr, "remote", "r", "", "Remote address to listen on")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshAddr, "ssh", "s", "", "SSH server address (host:port)")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshUser, "user", "u", "", "SSH username")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshPassword, "password", "p", "", "SSH password")
	reverseSshTunnelCmd.Flags().StringVarP(&tunnelFlags.sshKeyPath, "key", "k", "", "SSH private key path")
}
