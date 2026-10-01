package machine

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/tanq16/anbu/internal/awsx"
	"github.com/tanq16/anbu/internal/scaffold"
	"github.com/tanq16/anbu/internal/sshx"
	"github.com/tanq16/anbu/internal/vault"
)

const (
	probeCommand = "cloud-init status; test -f /home/ubuntu/.sharingan_bootstrap_done && echo bootstrap-done || echo bootstrap-pending"
	probeTimeout = 60 * time.Second
)

var ErrSSH = errors.New("ssh failed")

func Endpoint(ctx context.Context, c *awsx.Clients, keys *vault.Store, name string) (sshx.Endpoint, error) {
	inst, err := requireInstance(ctx, c, name)
	if err != nil {
		return sshx.Endpoint{}, err
	}
	if inst.PublicIP == "" {
		return sshx.Endpoint{}, &NotRunningError{Name: name, State: inst.State}
	}
	sec, err := keys.Get(scaffold.KeyName(c.Account, c.Region))
	if err != nil {
		return sshx.Endpoint{}, fmt.Errorf("scaffold key %s: %w", scaffold.KeyName(c.Account, c.Region), err)
	}
	signer, err := vault.Signer(sec)
	if err != nil {
		return sshx.Endpoint{}, err
	}
	return sshx.Endpoint{
		Addr:   net.JoinHostPort(inst.PublicIP, "22"),
		User:   SSHUser,
		Signer: signer,
		Alias:  c.HostKeyAlias(name),
	}, nil
}

func Status(ctx context.Context, c *awsx.Clients, keys *vault.Store, kh *sshx.KnownHosts, name string) (string, error) {
	ep, err := Endpoint(ctx, c, keys, name)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	client, err := sshx.Dial(ctx, ep, kh)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSSH, err)
	}
	defer client.Close()
	res, err := sshx.Exec(ctx, client, probeCommand)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSSH, err)
	}
	return strings.TrimSpace(res.Stdout + res.Stderr), nil
}
