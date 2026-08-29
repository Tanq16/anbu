package anbuNetwork

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	u "github.com/tanq16/anbu/utils"
)

type closerSet struct {
	mu      sync.Mutex
	closers []io.Closer
}

func (s *closerSet) add(c io.Closer) {
	s.mu.Lock()
	s.closers = append(s.closers, c)
	s.mu.Unlock()
}

func (s *closerSet) closeAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.closers {
		_ = c.Close()
	}
	s.closers = nil
}

func TCPTunnel(ctx context.Context, localAddr, remoteAddr string, useTLS, insecureSkipVerify bool) error {
	u.PrintInfo(fmt.Sprintf("TCP tunnel %s %s %s", localAddr, u.StyleSymbols["arrow"], remoteAddr))

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", localAddr, err)
	}
	defer listener.Close()
	u.PrintInfo(fmt.Sprintf("Listening on %s", localAddr))
	if useTLS {
		u.PrintStream("Using TLS for remote connections")
	}

	var live closerSet
	go func() {
		<-ctx.Done()
		u.PrintInfo("TCP tunnel stopped gracefully")
		listener.Close()
		live.closeAll()
	}()

	var activeConns sync.WaitGroup
	sem := make(chan struct{}, 100)
	for {
		select {
		case <-ctx.Done():
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
				u.PrintError("Failed to accept connection", err)
				continue
			}

			select {
			case sem <- struct{}{}:
			default:
				u.PrintWarn("Connection limit reached, rejecting", nil)
				localConn.Close()
				continue
			}
			live.add(localConn)
			activeConns.Go(func() {
				defer func() { <-sem }()
				defer localConn.Close()
				u.PrintInfo(fmt.Sprintf("New connection from %s", localConn.RemoteAddr()))

				select {
				case <-ctx.Done():
					return
				default:
				}

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
					u.PrintError(fmt.Sprintf("Failed to connect to remote %s", remoteAddr), err)
					return
				}
				live.add(remoteConn)
				defer remoteConn.Close()
				u.PrintInfo(fmt.Sprintf("Connected to remote %s", remoteAddr))

				var wg sync.WaitGroup
				wg.Go(func() {
					n, err := io.Copy(remoteConn, localConn)
					if err != nil && err != io.EOF {
						u.PrintError("Error copying data to remote", err)
					}
					u.PrintStream(fmt.Sprintf("%s Sent %d bytes to remote", u.StyleSymbols["arrow"], n))
				})
				wg.Go(func() {
					n, err := io.Copy(localConn, remoteConn)
					if err != nil && err != io.EOF {
						u.PrintError("Error copying data from remote", err)
					}
					u.PrintStream(fmt.Sprintf("← Received %d bytes from remote", n))
				})
				wg.Wait()
				u.PrintSuccess(fmt.Sprintf("Connection closed from %s", localConn.RemoteAddr()))
			})
		}
	}
}
