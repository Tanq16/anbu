package anbuNetwork

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/tanq16/anbu/utils"
)

func TCPTunnel(localAddr, remoteAddr string, useTLS, insecureSkipVerify bool) error {
	logger := utils.NewManager(0)
	logger.StartDisplay()
	defer logger.StopDisplay()
	funcID := logger.Register("tcp-tunnel")
	logger.SetMessage(funcID, fmt.Sprintf("TCP tunnel %s → %s", localAddr, remoteAddr))

	// Listen on the local address
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		logger.ReportError(funcID, fmt.Errorf("failed to listen on %s: %w", localAddr, err))
		return err
	}
	defer listener.Close()
	logger.AddStreamLine(funcID, fmt.Sprintf("Listening on %s", localAddr))
	if useTLS {
		logger.AddStreamLine(funcID, "Using TLS for remote connections")
	}

	// For graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	var activeConns sync.WaitGroup
	go func() {
		<-sigChan
		close(done)
		logger.Complete(funcID, "TCP tunnel stopped gracefully")
		listener.Close()
	}()

	// Continue accepting new connections until explicitly stopped
	for {
		select {
		case <-done:
			activeConns.Wait()
			return nil
		default:
			listener.(*net.TCPListener).SetDeadline(time.Now().Add(2 * time.Second))
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

				// Connect to remote
				var remoteConn net.Conn
				if useTLS {
					tlsConfig := &tls.Config{
						InsecureSkipVerify: insecureSkipVerify,
					}
					remoteConn, err = tls.Dial("tcp", remoteAddr, tlsConfig)
				} else {
					remoteConn, err = net.Dial("tcp", remoteAddr)
				}
				if err != nil {
					logger.AddStreamLine(funcID, fmt.Sprintf("Failed to connect to remote %s: %v", remoteAddr, err))
					return
				}
				defer remoteConn.Close()
				logger.AddStreamLine(funcID, fmt.Sprintf("Connected to remote %s", remoteAddr))

				// Copy data bidirectionally
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					// Local to Remote
					n, err := io.Copy(remoteConn, localConn)
					if err != nil && err != io.EOF {
						logger.AddStreamLine(funcID, fmt.Sprintf("Error copying data to remote: %v", err))
					}
					logger.AddStreamLine(funcID, fmt.Sprintf("→ Sent %d bytes to remote", n))
				}()
				go func() {
					defer wg.Done()
					// Remote to Local
					n, err := io.Copy(localConn, remoteConn)
					if err != nil && err != io.EOF {
						logger.AddStreamLine(funcID, fmt.Sprintf("Error copying data from remote: %v", err))
					}
					logger.AddStreamLine(funcID, fmt.Sprintf("← Received %d bytes from remote", n))
				}()
				wg.Wait()
				logger.AddStreamLine(funcID, fmt.Sprintf("Connection closed from %s", localConn.RemoteAddr()))
			}()
		}
	}
}

func ReverseTCPTunnel(localAddr, remoteAddr string, useTLS, insecureSkipVerify bool) error {
	logger := utils.NewManager(0)
	logger.StartDisplay()
	defer logger.StopDisplay()
	funcID := logger.Register("reverse-tcp-tunnel")
	logger.SetMessage(funcID, fmt.Sprintf("Reverse TCP tunnel %s ← %s", localAddr, remoteAddr))

	// For graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	var activeConns sync.WaitGroup
	go func() {
		<-sigChan
		close(done)
		logger.Complete(funcID, "Reverse TCP tunnel stopped gracefully")
	}()
	logger.AddStreamLine(funcID, fmt.Sprintf("Connecting to remote %s", remoteAddr))

	// Continue connecting to the remote server until explicitly stopped
	for {
		select {
		case <-done:
			activeConns.Wait()
			return nil
		default:
			var remoteConn net.Conn
			var err error
			if useTLS {
				tlsConfig := &tls.Config{
					InsecureSkipVerify: insecureSkipVerify,
				}
				remoteConn, err = tls.Dial("tcp", remoteAddr, tlsConfig)
			} else {
				remoteConn, err = net.Dial("tcp", remoteAddr)
			}
			if err != nil {
				logger.AddStreamLine(funcID, fmt.Sprintf("Failed to connect to remote at %s: %v", remoteAddr, err))
				select {
				case <-done:
					activeConns.Wait()
					return nil
				case <-time.After(5 * time.Second):
					continue
				}
			}

			// Handle the connection in a goroutine
			activeConns.Add(1)
			go func(remoteConn net.Conn) {
				defer activeConns.Done()
				defer remoteConn.Close()
				logger.AddStreamLine(funcID, fmt.Sprintf("Connected to remote %s", remoteAddr))
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
					// Local to Remote
					n, err := io.Copy(remoteConn, localConn)
					if err != nil && err != io.EOF {
						logger.AddStreamLine(funcID, fmt.Sprintf("Error copying data to remote: %v", err))
					}
					logger.AddStreamLine(funcID, fmt.Sprintf("→ Sent %d bytes to remote", n))
				}()
				go func() {
					defer wg.Done()
					// Remote to Local
					n, err := io.Copy(localConn, remoteConn)
					if err != nil && err != io.EOF {
						logger.AddStreamLine(funcID, fmt.Sprintf("Error copying data from remote: %v", err))
					}
					logger.AddStreamLine(funcID, fmt.Sprintf("← Received %d bytes from remote", n))
				}()
				wg.Wait()
				logger.AddStreamLine(funcID, "Connection closed")
			}(remoteConn)
		}
	}
}
