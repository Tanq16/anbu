package wgproxy

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
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

func handleSOCKS5(ctx context.Context, client net.Conn, dial dialFunc) error {
	defer client.Close()
	reader := bufio.NewReader(client)

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
	return tunnelTCP(ctx, src, client, target, dial, func(net.Conn) error {
		_, err := client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return err
	})
}

func tunnelTCP(ctx context.Context, src, client net.Conn, target string, dial dialFunc, reply func(net.Conn) error) error {
	log.Debug().Str("target", target).Msg("proxy dial")
	remote, err := dial(ctx, "tcp", target)
	if err != nil {
		_, _ = client.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
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
