package anbuNetwork

// import (
// 	"fmt"
// 	"io"
// 	"net"
// 	"os"
// 	"os/signal"
// 	"sync"
// 	"syscall"
// 	"time"

// 	"github.com/tanq16/anbu/utils"
// 	"golang.org/x/crypto/ssh"
// )

// func SSHTunnel(localAddr, remoteAddr, sshAddr, user string, authMethods []ssh.AuthMethod) error {
// 	config := &ssh.ClientConfig{
// 		User:            user,
// 		Auth:            authMethods,
// 		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
// 		Timeout:         30 * time.Second,
// 	}
// 	// Connect to SSH server
// 	sshClient, err := ssh.Dial("tcp", sshAddr, config)
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to SSH server: %w", err)
// 	}
// 	defer sshClient.Close()
// 	// Listen on local address
// 	listener, err := net.Listen("tcp", localAddr)
// 	if err != nil {
// 		return fmt.Errorf("failed to listen on %s: %w", localAddr, err)
// 	}
// 	defer listener.Close()

// 	logger.Debug().Str("local", localAddr).Str("remote", remoteAddr).Str("ssh", sshAddr).Msg("SSH tunnel started")
// 	fmt.Println(utils.OutSuccess(fmt.Sprintf("SSH tunnel %s ➜ %s via %s", localAddr, remoteAddr, sshAddr)))
// 	fmt.Println(utils.OutWarning("Press Ctrl+C to stop"))

// 	// For graceful shutdown
// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
// 	done := make(chan struct{})
// 	var activeConns sync.WaitGroup
// 	go func() {
// 		<-sigChan
// 		close(done)
// 		listener.Close()
// 		sshClient.Close()
// 		logger.Debug().Msg("Shutting down tunnel")
// 	}()

// 	for {
// 		select {
// 		case <-done:
// 			activeConns.Wait()
// 			return nil
// 		default:
// 			listener.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second))
// 			localConn, err := listener.Accept()
// 			if err != nil {
// 				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
// 					continue
// 				}
// 				if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
// 					return nil
// 				}
// 				logger.Debug().Err(err).Msg("Failed to accept connection")
// 				continue
// 			}

// 			// Handle the connection in a goroutine
// 			activeConns.Add(1)
// 			go func(localConn net.Conn) {
// 				defer activeConns.Done()
// 				defer localConn.Close()
// 				// Connect to remote through SSH
// 				remoteConn, err := sshClient.Dial("tcp", remoteAddr)
// 				if err != nil {
// 					logger.Error().Err(err).Str("remote", remoteAddr).Msg("Failed to connect to remote via SSH")
// 					fmt.Println(utils.OutError("Failed to connect to remote via SSH"))
// 					return
// 				}
// 				defer remoteConn.Close()
// 				logger.Debug().Str("remote", remoteAddr).Msg("New SSH tunnel connection established")

// 				// Copy data bidirectionally
// 				var wg sync.WaitGroup
// 				wg.Add(2)
// 				go func() {
// 					defer wg.Done()
// 					io.Copy(remoteConn, localConn)
// 				}()
// 				go func() {
// 					defer wg.Done()
// 					io.Copy(localConn, remoteConn)
// 				}()
// 				wg.Wait()
// 				logger.Debug().Msg("SSH tunnel connection closed")
// 			}(localConn)
// 		}
// 	}
// }

// func ReverseSSHTunnel(localAddr, remoteAddr, sshAddr, user string, authMethods []ssh.AuthMethod) error {
// 	config := &ssh.ClientConfig{
// 		User:            user,
// 		Auth:            authMethods,
// 		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
// 		Timeout:         30 * time.Second,
// 	}
// 	// Connect to SSH server
// 	sshClient, err := ssh.Dial("tcp", sshAddr, config)
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to SSH server: %w", err)
// 	}
// 	defer sshClient.Close()
// 	// Start listening on remote
// 	listener, err := sshClient.Listen("tcp", remoteAddr)
// 	if err != nil {
// 		return fmt.Errorf("failed to listen on remote address %s: %w", remoteAddr, err)
// 	}
// 	defer listener.Close()

// 	logger.Debug().Str("local", localAddr).Str("remote", remoteAddr).Str("ssh", sshAddr).Msg("Reverse SSH tunnel started")
// 	fmt.Println(utils.OutSuccess(fmt.Sprintf("Reverse SSH tunnel %s ➜ %s via %s", remoteAddr, localAddr, sshAddr)))
// 	fmt.Println(utils.OutWarning("Press Ctrl+C to stop"))

// 	// For graceful shutdown
// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
// 	done := make(chan struct{})
// 	var activeConns sync.WaitGroup
// 	go func() {
// 		<-sigChan
// 		close(done)
// 		listener.Close()
// 		sshClient.Close()
// 		logger.Debug().Msg("Shutting down tunnel")
// 	}()

// 	for {
// 		select {
// 		case <-done:
// 			activeConns.Wait()
// 			return nil
// 		default:
// 			if tcpListener, ok := listener.(*net.TCPListener); ok {
// 				tcpListener.SetDeadline(time.Now().Add(time.Second))
// 			}
// 			remoteConn, err := listener.Accept()
// 			if err != nil {
// 				if netErr, ok := err.(net.Error); ok {
// 					if netErr.Timeout() { // Check for timeout
// 						continue
// 					} else { // Potentially shutdown error
// 						return nil
// 					}
// 				}
// 				logger.Debug().Err(err).Msg("Failed to accept connection")
// 				continue
// 			}

// 			// Handle connection in a goroutine
// 			activeConns.Add(1)
// 			go func(remoteConn net.Conn) {
// 				defer activeConns.Done()
// 				defer remoteConn.Close()
// 				// Connect to the local service
// 				localConn, err := net.Dial("tcp", localAddr)
// 				if err != nil {
// 					logger.Error().Err(err).Str("local", localAddr).Msg("Failed to connect to local service")
// 					fmt.Println(utils.OutError(fmt.Sprintf("Failed to connect to local service at %s", localAddr)))
// 					return
// 				}
// 				defer localConn.Close()
// 				logger.Debug().Str("local", localAddr).Msg("New reverse SSH tunnel connection established")

// 				// Copy data bidirectionally
// 				var wg sync.WaitGroup
// 				wg.Add(2)
// 				go func() {
// 					defer wg.Done()
// 					io.Copy(localConn, remoteConn)
// 				}()
// 				go func() {
// 					defer wg.Done()
// 					io.Copy(remoteConn, localConn)
// 				}()
// 				wg.Wait()
// 				logger.Debug().Msg("Reverse SSH tunnel connection closed")
// 			}(remoteConn)
// 		}
// 	}
// }

// // Authentication helper functions

// func TunnelSSHPassword(password string) ssh.AuthMethod {
// 	return ssh.Password(password)
// }

// func TunnelSSHPrivateKey(keyPath string) (ssh.AuthMethod, error) {
// 	key, err := os.ReadFile(keyPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to read private key: %w", err)
// 	}
// 	signer, err := ssh.ParsePrivateKey(key)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to parse private key: %w", err)
// 	}
// 	return ssh.PublicKeys(signer), nil
// }
