package anbuNetwork

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tanq16/anbu/utils"
)

func TCPTunnel(localAddr, remoteAddr string, useTLS, insecureSkipVerify bool) error {
	logger := utils.GetLogger("tunnel-tcp")
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
	go func() {
		<-sigChan
		close(done)
		listener.Close()
		fmt.Println(utils.OutSuccess("TCP tunnel stopped gracefully"))
	}()

	for {
		select {
		case <-done:
			return nil
		default:
			listener.(*net.TCPListener).SetDeadline(time.Now().Add(2 * time.Second))
			conn, err := listener.Accept()
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
			go handleConnection(conn, remoteAddr, useTLS, insecureSkipVerify)
		}
	}
}

func handleConnection(localConn net.Conn, remoteAddr string, useTLS, insecureSkipVerify bool) {
	logger := utils.GetLogger("tunnel-tcp")
	defer localConn.Close()
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
		return
	}
	defer remoteConn.Close()
	logger.Debug().Str("remote", remoteAddr).Msg("New connection established to remote")

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
