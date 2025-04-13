package anbuNetwork

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tanq16/anbu/utils"
	"golang.org/x/crypto/ssh"
)

func SSHTunnel(localAddr, remoteAddr, sshAddr, user string, authMethods []ssh.AuthMethod) error {
	logger := utils.GetLogger("tunnel-ssh")
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}
	sshClient, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}
	defer sshClient.Close()

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", localAddr, err)
	}
	defer listener.Close()

	logger.Debug().Str("local", localAddr).Str("remote", remoteAddr).Str("ssh", sshAddr).Msg("SSH tunnel started")
	fmt.Println(utils.OutSuccess(fmt.Sprintf("SSH tunnel %s ➜ %s via %s", localAddr, remoteAddr, sshAddr)))
	fmt.Println(utils.OutWarning("Press Ctrl+C to stop"))

	// For graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		<-sigChan
		close(done)
		listener.Close()
		logger.Info().Msg("Shutting down tunnel")
	}()

	for {
		select {
		case <-done:
			return nil
		default:
			listener.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second))
			localConn, err := listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
					return nil
				}
				logger.Error().Err(err).Msg("Failed to accept connection")
				continue
			}
			go handleSSHConnection(localConn, remoteAddr, sshClient)
		}
	}
}

func handleSSHConnection(localConn net.Conn, remoteAddr string, sshClient *ssh.Client) {
	logger := utils.GetLogger("tunnel-ssh")
	defer localConn.Close()

	remoteConn, err := sshClient.Dial("tcp", remoteAddr)
	if err != nil {
		logger.Debug().Err(err).Str("remote", remoteAddr).Msg("Failed to connect to remote via SSH")
		fmt.Println(utils.OutError("Failed to connect to remote via SSH"))
		return
	}
	defer remoteConn.Close()
	logger.Debug().Str("remote", remoteAddr).Msg("New SSH tunnel connection established")

	// Copy data in both directions
	go func() {
		_, err := io.Copy(remoteConn, localConn)
		if err != nil && err != io.EOF {
			logger.Debug().Err(err).Msg("Error copying data to remote")
		}
	}()
	_, err = io.Copy(localConn, remoteConn)
	if err != nil && err != io.EOF {
		logger.Debug().Err(err).Msg("Error copying data from remote")
	}
}

func Password(password string) ssh.AuthMethod {
	return ssh.Password(password)
}

func PrivateKey(keyPath string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}
	return ssh.PublicKeys(signer), nil
}
