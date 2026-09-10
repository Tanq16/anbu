package wgproxy

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/rs/zerolog/log"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

type Server struct {
	cfg    Config
	device *device.Device
	tnet   *netstack.Net
}

func New(cfg Config) (*Server, error) {
	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	dev, tnet, err := setupTunnel(cfg)
	if err != nil {
		return nil, err
	}
	return &Server{cfg: cfg, device: dev, tnet: tnet}, nil
}

func (s *Server) Serve(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	stop := context.AfterFunc(ctx, func() {
		ln.Close()
	})
	defer stop()

	var wg sync.WaitGroup
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				wg.Wait()
				return nil
			}
			return err
		}
		wg.Go(func() {
			if err := handleSOCKS5(ctx, conn, s.dial); err != nil {
				log.Debug().Err(err).Msg("proxy connection")
			}
		})
	}
}

func (s *Server) dial(ctx context.Context, network, addr string) (net.Conn, error) {
	if s.cfg.DialTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.cfg.DialTimeout)
		defer cancel()
	}
	return s.tnet.DialContext(ctx, network, addr)
}

func (s *Server) Close() {
	if s.device != nil {
		s.device.Close()
		s.device = nil
	}
}
