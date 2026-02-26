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

	u "github.com/tanq16/anbu/internal/utils"
	"golang.org/x/crypto/ssh"
)

func SSHTunnel(localAddr, remoteAddr, sshAddr, user string, authMethods []ssh.AuthMethod) {
	u.PrintInfo(fmt.Sprintf("SSH tunnel %s %s %s via %s", localAddr, u.StyleSymbols["arrow"], remoteAddr, sshAddr))
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	u.PrintInfo(fmt.Sprintf("Connecting to SSH server at %s...", sshAddr))
	sshClient, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		u.PrintFatal(fmt.Sprintf("failed to connect to SSH server: %s", sshAddr), err)
	}
	defer sshClient.Close()
	u.PrintInfo(fmt.Sprintf("Connected to SSH server as %s", user))

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		u.PrintFatal(fmt.Sprintf("failed to listen on %s", localAddr), err)
	}
	defer listener.Close()
	u.PrintInfo(fmt.Sprintf("Listening on %s", localAddr))

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
				u.PrintWarn("Failed to accept connection", err)
				continue
			}

			activeConns.Add(1)
			go func() {
				defer activeConns.Done()
				defer localConn.Close()
				u.PrintInfo(fmt.Sprintf("New connection from %s", localConn.RemoteAddr()))
				remoteConn, err := sshClient.Dial("tcp", remoteAddr)
				if err != nil {
					u.PrintError("Failed to connect to remote via SSH", err)
					return
				}
				defer remoteConn.Close()
				u.PrintInfo(fmt.Sprintf("Connected to remote %s via SSH", remoteAddr))

				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					n, err := io.Copy(remoteConn, localConn)
					if err != nil && err != io.EOF {
						u.PrintError("Error copying data to remote", err)
					}
					u.PrintStream(fmt.Sprintf("%s Sent %d bytes to remote via SSH", u.StyleSymbols["arrow"], n))
				}()
				go func() {
					defer wg.Done()
					n, err := io.Copy(localConn, remoteConn)
					if err != nil && err != io.EOF {
						u.PrintError("Error copying data from remote", err)
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
	u.PrintInfo(fmt.Sprintf("Reverse SSH tunnel %s %s %s via %s", remoteAddr, u.StyleSymbols["arrow"], localAddr, sshAddr))
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	u.PrintInfo(fmt.Sprintf("Connecting to SSH server at %s...", sshAddr))
	sshClient, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		u.PrintFatal("failed to connect to SSH server", err)
	}
	defer sshClient.Close()
	u.PrintInfo(fmt.Sprintf("Connected to SSH server as %s", user))

	u.PrintInfo(fmt.Sprintf("Setting up listener on remote address %s", remoteAddr))
	listener, err := sshClient.Listen("tcp", remoteAddr)
	if err != nil {
		u.PrintFatal(fmt.Sprintf("failed to listen on remote address %s", remoteAddr), err)
	}
	defer listener.Close()
	u.PrintInfo(fmt.Sprintf("Listening on remote address %s", remoteAddr))

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
				u.PrintWarn("Failed to accept connection", err)
				continue
			}

			activeConns.Add(1)
			go func() {
				defer activeConns.Done()
				defer remoteConn.Close()
				u.PrintInfo(fmt.Sprintf("New connection from remote %s", remoteConn.RemoteAddr()))

				localConn, err := net.Dial("tcp", localAddr)
				if err != nil {
					u.PrintError(fmt.Sprintf("Failed to connect to local service at %s", localAddr), err)
					return
				}
				defer localConn.Close()
				u.PrintInfo(fmt.Sprintf("Connected to local service %s", localAddr))

				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					n, err := io.Copy(remoteConn, localConn)
					if err != nil && err != io.EOF {
						u.PrintError("Error copying data to remote", err)
					}
					u.PrintStream(fmt.Sprintf("%s Sent %d bytes to remote", u.StyleSymbols["arrow"], n))
				}()
				go func() {
					defer wg.Done()
					n, err := io.Copy(localConn, remoteConn)
					if err != nil && err != io.EOF {
						u.PrintError("Error copying data from remote", err)
					}
					u.PrintStream(fmt.Sprintf("← Received %d bytes from remote", n))
				}()
				wg.Wait()
				u.PrintInfo(fmt.Sprintf("Connection closed from remote %s", remoteConn.RemoteAddr()))
			}()
		}
	}
}

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
