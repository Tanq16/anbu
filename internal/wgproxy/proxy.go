package wgproxy

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strconv"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type dialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

type prefixConn struct {
	net.Conn
	r io.Reader
}

func (c prefixConn) Read(p []byte) (int, error) {
	return c.r.Read(p)
}

func (c prefixConn) CloseWrite() error {
	type closeWriter interface {
		CloseWrite() error
	}
	if cw, ok := c.Conn.(closeWriter); ok {
		return cw.CloseWrite()
	}
	return c.Close()
}

func handleHybrid(ctx context.Context, client net.Conn, dial dialFunc) error {
	defer client.Close()
	reader := bufio.NewReader(client)
	peek, err := reader.Peek(1)
	if err != nil {
		return err
	}
	switch peek[0] {
	case 0x05:
		return handleSOCKS5(ctx, client, reader, dial)
	case 0x04:
		return errors.New("SOCKS4 is not supported")
	default:
		return handleHTTP(ctx, client, reader, dial)
	}
}

func handleHTTP(ctx context.Context, client net.Conn, reader *bufio.Reader, dial dialFunc) error {
	req, err := http.ReadRequest(reader)
	if err != nil {
		return err
	}
	defer req.Body.Close()

	src := prefixConn{Conn: client, r: reader}
	if req.Method == http.MethodConnect {
		target := req.Host
		if _, _, splitErr := net.SplitHostPort(target); splitErr != nil {
			target = net.JoinHostPort(target, "443")
		}
		return tunnelTCP(ctx, src, client, target, dial, httpDialFail, func(net.Conn) error {
			_, err := client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
			return err
		})
	}

	target := req.URL.Host
	if target == "" {
		target = req.Host
	}
	if _, _, splitErr := net.SplitHostPort(target); splitErr != nil {
		target = net.JoinHostPort(target, "80")
	}
	return tunnelTCP(ctx, src, client, target, dial, httpDialFail, func(remote net.Conn) error {
		req.RequestURI = ""
		req.Header.Del("Proxy-Connection")
		req.Header.Del("Proxy-Authenticate")
		req.Header.Del("Proxy-Authorization")
		return req.Write(remote)
	})
}

func handleSOCKS5(ctx context.Context, client net.Conn, reader *bufio.Reader, dial dialFunc) error {
	header := make([]byte, 2)
	if _, err := io.ReadFull(reader, header); err != nil {
		return err
	}
	if header[0] != 0x05 {
		return errors.New("unsupported SOCKS version")
	}
	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(reader, methods); err != nil {
		return err
	}
	if !slices.Contains(methods, 0x00) {
		_, _ = client.Write([]byte{0x05, 0xFF})
		return errors.New("SOCKS5 client requires authentication")
	}
	if _, err := client.Write([]byte{0x05, 0x00}); err != nil {
		return err
	}

	req := make([]byte, 4)
	if _, err := io.ReadFull(reader, req); err != nil {
		return err
	}
	if req[1] != 0x01 {
		_, _ = client.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return errors.New("only SOCKS5 CONNECT is supported")
	}

	var host string
	switch req[3] {
	case 0x01:
		ip := make([]byte, 4)
		if _, err := io.ReadFull(reader, ip); err != nil {
			return err
		}
		host = net.IP(ip).String()
	case 0x03:
		length, err := reader.ReadByte()
		if err != nil {
			return err
		}
		domain := make([]byte, int(length))
		if _, err := io.ReadFull(reader, domain); err != nil {
			return err
		}
		host = string(domain)
	case 0x04:
		ip := make([]byte, 16)
		if _, err := io.ReadFull(reader, ip); err != nil {
			return err
		}
		host = net.IP(ip).String()
	default:
		_, _ = client.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return errors.New("unsupported SOCKS5 address type")
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(reader, portBuf); err != nil {
		return err
	}
	target := net.JoinHostPort(host, strconv.Itoa(int(binary.BigEndian.Uint16(portBuf))))
	src := prefixConn{Conn: client, r: reader}
	return tunnelTCP(ctx, src, client, target, dial, socksDialFail, func(net.Conn) error {
		_, err := client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return err
	})
}

var (
	httpDialFail  = []byte("HTTP/1.1 502 Bad Gateway\r\n\r\n")
	socksDialFail = []byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
)

func tunnelTCP(ctx context.Context, src, client net.Conn, target string, dial dialFunc, failReply []byte, reply func(net.Conn) error) error {
	log.Debug().Str("target", target).Msg("proxy dial")
	remote, err := dial(ctx, "tcp", target)
	if err != nil {
		_, _ = client.Write(failReply)
		return fmt.Errorf("dial %s: %w", target, err)
	}
	defer remote.Close()
	if err := reply(remote); err != nil {
		return err
	}
	return pipeConnections(ctx, src, remote)
}

func pipeConnections(ctx context.Context, a, b net.Conn) error {
	g, gctx := errgroup.WithContext(ctx)
	stop := context.AfterFunc(gctx, func() {
		a.Close()
		b.Close()
	})
	defer stop()
	g.Go(func() error {
		_, err := io.Copy(b, a)
		closeWrite(b)
		return ignorePipeErr(err)
	})
	g.Go(func() error {
		_, err := io.Copy(a, b)
		closeWrite(a)
		return ignorePipeErr(err)
	})
	return g.Wait()
}

func closeWrite(c net.Conn) {
	type closeWriter interface {
		CloseWrite() error
	}
	if cw, ok := c.(closeWriter); ok {
		_ = cw.CloseWrite()
		return
	}
	c.Close()
}

func ignorePipeErr(err error) error {
	if err == nil || errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}
