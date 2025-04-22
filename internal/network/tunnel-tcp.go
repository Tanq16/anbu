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
	// Listen on the local address
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", localAddr, err)
	}
	defer listener.Close()
	logger.Debug().Str("local", localAddr).Str("remote", remoteAddr).Bool("tls", useTLS).Msg("TCP tunnel started")
	fmt.Println(utils.OutSuccess(fmt.Sprintf("TCP tunnel %s ➜ %s", localAddr, remoteAddr)))
	fmt.Println(utils.OutWarning("Press Ctrl+C to stop"))

	// For graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	var activeConns sync.WaitGroup
	go func() {
		<-sigChan
		close(done)
		listener.Close()
		activeConns.Wait()
		fmt.Println(utils.OutSuccess("TCP tunnel stopped gracefully"))
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
				logger.Debug().Err(err).Msg("Failed to accept connection")
				fmt.Println(utils.OutError("Failed to accept connection"))
				continue
			}

			// Handle the connection in a goroutine
			activeConns.Add(1)
			go func() {
				defer activeConns.Done()
				defer localConn.Close()
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
					logger.Error().Err(err).Str("remote", remoteAddr).Msg("Failed to connect to remote")
					return
				}
				defer remoteConn.Close()
				logger.Debug().Str("remote", remoteAddr).Msg("New connection established to remote")
				// Copy data bidirectionally
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					io.Copy(remoteConn, localConn)
				}()
				go func() {
					defer wg.Done()
					io.Copy(localConn, remoteConn)
				}()
				wg.Wait()
				logger.Debug().Msg("Connection closed")
			}()
		}
	}
}

func ReverseTCPTunnel(localAddr, remoteAddr string, useTLS, insecureSkipVerify bool) error {
	// For graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	var activeConns sync.WaitGroup
	go func() {
		<-sigChan
		close(done)
		fmt.Println(utils.OutSuccess("Reverse TCP tunnel stopping gracefully"))
	}()

	logger.Debug().Str("local", localAddr).Str("remote", remoteAddr).Bool("tls", useTLS).Msg("Reverse TCP tunnel starting")
	fmt.Println(utils.OutSuccess(fmt.Sprintf("Reverse TCP tunnel %s ➜ %s", remoteAddr, localAddr)))
	fmt.Println(utils.OutWarning("Press Ctrl+C to stop"))

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
				logger.Error().Err(err).Str("remote", remoteAddr).Msg("Failed to connect to remote")
				fmt.Println(utils.OutError(fmt.Sprintf("Failed to connect to remote at %s: %v", remoteAddr, err)))
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
				// Connect to the local service
				localConn, err := net.Dial("tcp", localAddr)
				if err != nil {
					logger.Error().Err(err).Str("local", localAddr).Msg("Failed to connect to local service")
					fmt.Println(utils.OutError(fmt.Sprintf("Failed to connect to local service at %s: %v", localAddr, err)))
					return
				}
				defer localConn.Close()
				logger.Debug().Str("local", localAddr).Str("remote", remoteAddr).Msg("Tunnel connection established")

				// Copy data bidirectionally
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					io.Copy(localConn, remoteConn)
				}()
				go func() {
					defer wg.Done()
					io.Copy(remoteConn, localConn)
				}()
				wg.Wait()
				logger.Debug().Msg("Connection closed")
			}(remoteConn)
		}
	}
}
