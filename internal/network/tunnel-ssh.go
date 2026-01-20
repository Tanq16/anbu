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

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
	"golang.org/x/crypto/ssh"
)

func SSHTunnel(localAddr, remoteAddr, sshAddr, user string, authMethods []ssh.AuthMethod) {
	u.PrintInfo(fmt.Sprintf("SSH tunnel %s → %s via %s", localAddr, remoteAddr, sshAddr))
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect to SSH server
	u.PrintInfo(fmt.Sprintf("Connecting to SSH server at %s...", sshAddr))
	sshClient, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		u.PrintError(fmt.Sprintf("failed to connect to SSH server: %s", sshAddr))
		log.Debug().Err(err).Msgf("failed to connect to SSH server: %s", sshAddr)
		os.Exit(1)
	}
	defer sshClient.Close()
	u.PrintInfo(fmt.Sprintf("Connected to SSH server as %s", user))

	// Listen on local address
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		u.PrintError(fmt.Sprintf("failed to listen on %s", localAddr))
		log.Debug().Err(err).Msgf("failed to listen on %s", localAddr)
		os.Exit(1)
	}
	defer listener.Close()
	u.PrintInfo(fmt.Sprintf("Listening on %s", localAddr))

	// For graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	var activeConns sync.WaitGroup
	go func() {
		<-sigChan
		close(done)
		u.PrintInfo("SSH tunnel stopped gracefully")
		listener.Close()
		sshClient.Close()
	}()

	for {
		select {
		case <-done:
			activeConns.Wait()
			return
		default:
			listener.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second))
			localConn, err := listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
					return
				}
				u.PrintWarning("Failed to accept connection")
				log.Debug().Err(err).Msg("Failed to accept connection")
				continue
			}

			// Handle the connection in a goroutine
			activeConns.Add(1)
			go func() {
				defer activeConns.Done()
				defer localConn.Close()
				u.PrintInfo(fmt.Sprintf("New connection from %s", localConn.RemoteAddr()))
				// Connect to remote through SSH
				remoteConn, err := sshClient.Dial("tcp", remoteAddr)
				if err != nil {
					u.PrintError("Failed to connect to remote via SSH")
					log.Debug().Err(err).Msg("Failed to connect to remote via SSH")
					return
				}
				defer remoteConn.Close()
				u.PrintInfo(fmt.Sprintf("Connected to remote %s via SSH", remoteAddr))

				// Copy data bidirectionally
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					// Local to Remote (through SSH)
					n, err := io.Copy(remoteConn, localConn)
					if err != nil && err != io.EOF {
						u.PrintError("Error copying data to remote")
						log.Debug().Err(err).Msg("Error copying data to remote")
					}
					u.PrintStream(fmt.Sprintf("→ Sent %d bytes to remote via SSH", n))
				}()
				go func() {
					defer wg.Done()
					// Remote to Local (through SSH)
					n, err := io.Copy(localConn, remoteConn)
					if err != nil && err != io.EOF {
						u.PrintError("Error copying data from remote")
						log.Debug().Err(err).Msg("Error copying data from remote")
					}
					u.PrintStream(fmt.Sprintf("← Received %d bytes from remote via SSH", n))
				}()
				wg.Wait()
				u.PrintInfo(fmt.Sprintf("Connection closed from %s", localConn.RemoteAddr()))
			}()
		}
	}
}

func ReverseSSHTunnel(localAddr, remoteAddr, sshAddr, user string, authMethods []ssh.AuthMethod) {
	u.PrintInfo(fmt.Sprintf("Reverse SSH tunnel %s ← %s via %s", localAddr, remoteAddr, sshAddr))
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect to SSH server
	u.PrintInfo(fmt.Sprintf("Connecting to SSH server at %s...", sshAddr))
	sshClient, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		u.PrintError("failed to connect to SSH server")
		log.Debug().Err(err).Msg("failed to connect to SSH server")
		os.Exit(1)
	}
	defer sshClient.Close()
	u.PrintInfo(fmt.Sprintf("Connected to SSH server as %s", user))

	// Start listening on remote
	u.PrintInfo(fmt.Sprintf("Setting up listener on remote address %s", remoteAddr))
	listener, err := sshClient.Listen("tcp", remoteAddr)
	if err != nil {
		u.PrintError(fmt.Sprintf("failed to listen on remote address %s", remoteAddr))
		log.Debug().Err(err).Msgf("failed to listen on remote address %s", remoteAddr)
		os.Exit(1)
	}
	defer listener.Close()
	u.PrintInfo(fmt.Sprintf("Listening on remote address %s", remoteAddr))

	// For graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	var activeConns sync.WaitGroup
	go func() {
		<-sigChan
		close(done)
		u.PrintInfo("Reverse SSH tunnel stopped gracefully")
		listener.Close()
		sshClient.Close()
	}()

	for {
		select {
		case <-done:
			activeConns.Wait()
			return
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
						return
					}
				}
				u.PrintWarning("Failed to accept connection")
				log.Debug().Err(err).Msg("Failed to accept connection")
				continue
			}

			// Handle connection in a goroutine
			activeConns.Add(1)
			go func() {
				defer activeConns.Done()
				defer remoteConn.Close()
				u.PrintInfo(fmt.Sprintf("New connection from remote %s", remoteConn.RemoteAddr()))

				// Connect to the local service
				localConn, err := net.Dial("tcp", localAddr)
				if err != nil {
					u.PrintError(fmt.Sprintf("Failed to connect to local service at %s", localAddr))
					log.Debug().Err(err).Msgf("Failed to connect to local service at %s", localAddr)
					return
				}
				defer localConn.Close()
				u.PrintInfo(fmt.Sprintf("Connected to local service %s", localAddr))

				// Copy data bidirectionally
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					// Local to Remote (through SSH)
					n, err := io.Copy(remoteConn, localConn)
					if err != nil && err != io.EOF {
						u.PrintError("Error copying data to remote")
						log.Debug().Err(err).Msg("Error copying data to remote")
					}
					u.PrintStream(fmt.Sprintf("→ Sent %d bytes to remote", n))
				}()
				go func() {
					defer wg.Done()
					// Remote to Local (through SSH)
					n, err := io.Copy(localConn, remoteConn)
					if err != nil && err != io.EOF {
						u.PrintError("Error copying data from remote")
						log.Debug().Err(err).Msg("Error copying data from remote")
					}
					u.PrintStream(fmt.Sprintf("← Received %d bytes from remote", n))
				}()
				wg.Wait()
				u.PrintInfo("Connection closed")
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
