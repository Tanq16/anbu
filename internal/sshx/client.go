package sshx

import (
	"context"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

func Dial(ctx context.Context, ep Endpoint, kh *KnownHosts) (*ssh.Client, error) {
	cfg := &ssh.ClientConfig{
		User:            ep.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(ep.Signer)},
		HostKeyCallback: kh.Callback(ep.Alias),
		Timeout:         15 * time.Second,
	}
	conn, err := (&net.Dialer{Timeout: cfg.Timeout}).DialContext(ctx, "tcp", ep.Addr)
	if err != nil {
		return nil, err
	}
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	conn.SetDeadline(time.Now().Add(cfg.Timeout))
	c, chans, reqs, err := ssh.NewClientConn(conn, ep.Addr, cfg)
	stop()
	if err != nil {
		conn.Close()
		return nil, err
	}
	conn.SetDeadline(time.Time{})
	return ssh.NewClient(c, chans, reqs), nil
}
