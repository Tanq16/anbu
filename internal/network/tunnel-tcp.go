package anbuNetwork

import (
	"crypto/tls"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

func TCPTunnel(localAddr, remoteAddr string, useTLS, insecureSkipVerify bool) {
	log.Info().Msgf("TCP tunnel %s → %s", localAddr, remoteAddr)

	// Listen on the local address
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to listen on %s", localAddr)
	}
	defer listener.Close()
	log.Info().Msgf("Listening on %s", localAddr)
	if useTLS {
		log.Info().Msg("Using TLS for remote connections")
	}

	// For graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	var activeConns sync.WaitGroup
	go func() {
		<-sigChan
		close(done)
		log.Info().Msg("TCP tunnel stopped gracefully")
		listener.Close()
	}()

	// Continue accepting new connections until explicitly stopped
	for {
		select {
		case <-done:
			activeConns.Wait()
			return
		default:
			listener.(*net.TCPListener).SetDeadline(time.Now().Add(2 * time.Second))
			localConn, err := listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
					return
				}
				log.Error().Err(err).Msg("Failed to accept connection")
				continue
			}

			// Handle the connection in a goroutine
			activeConns.Add(1)
			go func() {
				defer activeConns.Done()
				defer localConn.Close()
				log.Info().Msgf("New connection from %s", localConn.RemoteAddr())

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
					log.Error().Err(err).Msgf("Failed to connect to remote %s", remoteAddr)
					return
				}
				defer remoteConn.Close()
				log.Info().Msgf("Connected to remote %s", remoteAddr)

				// Copy data bidirectionally
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					// Local to Remote
					n, err := io.Copy(remoteConn, localConn)
					if err != nil && err != io.EOF {
						log.Error().Err(err).Msg("Error copying data to remote")
					}
					log.Debug().Msgf("→ Sent %d bytes to remote", n)
				}()
				go func() {
					defer wg.Done()
					// Remote to Local
					n, err := io.Copy(localConn, remoteConn)
					if err != nil && err != io.EOF {
						log.Error().Err(err).Msg("Error copying data from remote")
					}
					log.Debug().Msgf("← Received %d bytes from remote", n)
				}()
				wg.Wait()
				log.Info().Msgf("Connection closed from %s", localConn.RemoteAddr())
			}()
		}
	}
}
