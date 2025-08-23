package anbuNetwork

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/tanq16/anbu/utils"
	"golang.org/x/crypto/ssh"
)

func SSHTunnel(localAddr, remoteAddr, sshAddr, user string, authMethods []ssh.AuthMethod) error {
	logger := utils.NewManager()
	logger.StartDisplay()
	defer logger.StopDisplay()
	funcID := logger.Register("ssh-tunnel")
	logger.SetMessage(funcID, fmt.Sprintf("SSH tunnel %s → %s via %s", localAddr, remoteAddr, sshAddr))

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect to SSH server
	logger.AddStreamLine(funcID, fmt.Sprintf("Connecting to SSH server at %s...", sshAddr))
	sshClient, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		logger.ReportError(funcID, fmt.Errorf("failed to connect to SSH server: %w", err))
		return err
	}
	defer sshClient.Close()
	logger.AddStreamLine(funcID, fmt.Sprintf("Connected to SSH server as %s", user))

	// Listen on local address
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		logger.ReportError(funcID, fmt.Errorf("failed to listen on %s: %w", localAddr, err))
		return err
	}
	defer listener.Close()
	logger.AddStreamLine(funcID, fmt.Sprintf("Listening on %s", localAddr))

	// For graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	var activeConns sync.WaitGroup
	go func() {
		<-sigChan
		close(done)
		logger.Complete(funcID, "SSH tunnel stopped gracefully")
		listener.Close()
		sshClient.Close()
	}()

	for {
		select {
		case <-done:
			activeConns.Wait()
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
				logger.AddStreamLine(funcID, fmt.Sprintf("Failed to accept connection: %v", err))
				continue
			}

			// Handle the connection in a goroutine
			activeConns.Add(1)
			go func() {
				defer activeConns.Done()
				defer localConn.Close()
				logger.AddStreamLine(funcID, fmt.Sprintf("New connection from %s", localConn.RemoteAddr()))
				// Connect to remote through SSH
				remoteConn, err := sshClient.Dial("tcp", remoteAddr)
				if err != nil {
					logger.AddStreamLine(funcID, fmt.Sprintf("Failed to connect to remote via SSH: %v", err))
					return
				}
				defer remoteConn.Close()
				logger.AddStreamLine(funcID, fmt.Sprintf("Connected to remote %s via SSH", remoteAddr))

				// Copy data bidirectionally
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()

					// Local to Remote (through SSH)
					n, err := io.Copy(remoteConn, localConn)
					if err != nil && err != io.EOF {
						logger.AddStreamLine(funcID, fmt.Sprintf("Error copying data to remote: %v", err))
					}
					logger.AddStreamLine(funcID, fmt.Sprintf("→ Sent %d bytes to remote via SSH", n))
				}()
				go func() {
					defer wg.Done()

					// Remote to Local (through SSH)
					n, err := io.Copy(localConn, remoteConn)
					if err != nil && err != io.EOF {
						logger.AddStreamLine(funcID, fmt.Sprintf("Error copying data from remote: %v", err))
					}
					logger.AddStreamLine(funcID, fmt.Sprintf("← Received %d bytes from remote via SSH", n))
				}()
				wg.Wait()
				logger.AddStreamLine(funcID, fmt.Sprintf("Connection closed from %s", localConn.RemoteAddr()))
			}()
		}
	}
}

func ReverseSSHTunnel(localAddr, remoteAddr, sshAddr, user string, authMethods []ssh.AuthMethod) error {
	logger := utils.NewManager()
	logger.StartDisplay()
	defer logger.StopDisplay()
	funcID := logger.Register("reverse-ssh-tunnel")
	logger.SetMessage(funcID, fmt.Sprintf("Reverse SSH tunnel %s ← %s via %s", localAddr, remoteAddr, sshAddr))

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect to SSH server
	logger.AddStreamLine(funcID, fmt.Sprintf("Connecting to SSH server at %s...", sshAddr))
	sshClient, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		logger.ReportError(funcID, fmt.Errorf("failed to connect to SSH server: %w", err))
		return err
	}
	defer sshClient.Close()
	logger.AddStreamLine(funcID, fmt.Sprintf("Connected to SSH server as %s", user))

	// Start listening on remote
	logger.AddStreamLine(funcID, fmt.Sprintf("Setting up listener on remote address %s", remoteAddr))
	listener, err := sshClient.Listen("tcp", remoteAddr)
	if err != nil {
		logger.ReportError(funcID, fmt.Errorf("failed to listen on remote address %s: %w", remoteAddr, err))
		return err
	}
	defer listener.Close()
	logger.AddStreamLine(funcID, fmt.Sprintf("Listening on remote address %s", remoteAddr))

	// For graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	var activeConns sync.WaitGroup
	go func() {
		<-sigChan
		close(done)
		logger.Complete(funcID, "Reverse SSH tunnel stopped gracefully")
		listener.Close()
		sshClient.Close()
	}()

	for {
		select {
		case <-done:
			activeConns.Wait()
			return nil
		default:
			if tcpListener, ok := listener.(*net.TCPListener); ok {
				tcpListener.SetDeadline(time.Now().Add(time.Second))
			}
			remoteConn, err := listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok {
					if netErr.Timeout() {
						continue
					} else {
						return nil
					}
				}
				logger.AddStreamLine(funcID, fmt.Sprintf("Failed to accept connection: %v", err))
				continue
			}

			// Handle connection in a goroutine
			activeConns.Add(1)
			go func() {
				defer activeConns.Done()
				defer remoteConn.Close()
				logger.AddStreamLine(funcID, fmt.Sprintf("New connection from remote %s", remoteConn.RemoteAddr()))

				// Connect to the local service
				localConn, err := net.Dial("tcp", localAddr)
				if err != nil {
					logger.AddStreamLine(funcID, fmt.Sprintf("Failed to connect to local service at %s: %v", localAddr, err))
					return
				}
				defer localConn.Close()
				logger.AddStreamLine(funcID, fmt.Sprintf("Connected to local service %s", localAddr))

				// Copy data bidirectionally
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()

					// Local to Remote (through SSH)
					n, err := io.Copy(remoteConn, localConn)
					if err != nil && err != io.EOF {
						logger.AddStreamLine(funcID, fmt.Sprintf("Error copying data to remote: %v", err))
					}
					logger.AddStreamLine(funcID, fmt.Sprintf("→ Sent %d bytes to remote", n))
				}()
				go func() {
					defer wg.Done()

					// Remote to Local (through SSH)
					n, err := io.Copy(localConn, remoteConn)
					if err != nil && err != io.EOF {
						logger.AddStreamLine(funcID, fmt.Sprintf("Error copying data from remote: %v", err))
					}
					logger.AddStreamLine(funcID, fmt.Sprintf("← Received %d bytes from remote", n))
				}()
				wg.Wait()
				logger.AddStreamLine(funcID, "Connection closed")
			}()
		}
	}
}

// Authentication helper functions

func TunnelSSHPassword(password string) ssh.AuthMethod {
	return ssh.Password(password)
}

func TunnelSSHPrivateKey(keyPath string) (ssh.AuthMethod, error) {
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
