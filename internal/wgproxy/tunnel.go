package wgproxy

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/rs/zerolog/log"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

func setupTunnel(cfg Config) (*device.Device, *netstack.Net, error) {
	privKey, err := NormalizeKey(cfg.PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid private key: %w", err)
	}
	peerKey, err := NormalizeKey(cfg.PeerPublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid peer public key: %w", err)
	}

	localIP, err := parseTunnelAddr(cfg.Address)
	if err != nil {
		return nil, nil, err
	}
	dnsIP, err := netip.ParseAddr(cfg.DNS)
	if err != nil {
		return nil, nil, err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", cfg.Endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve endpoint: %w", err)
	}
	endpoint := udpAddr.String()

	tunDev, tnet, err := netstack.CreateNetTUN([]netip.Addr{localIP}, []netip.Addr{dnsIP}, cfg.MTU)
	if err != nil {
		return nil, nil, err
	}

	logger := device.NewLogger(device.LogLevelSilent, "")
	if cfg.Debug {
		logger = &device.Logger{
			Verbosef: func(format string, args ...any) { log.Debug().Msgf(format, args...) },
			Errorf:   func(format string, args ...any) { log.Error().Msgf(format, args...) },
		}
	}

	dev := device.NewDevice(tunDev, conn.NewDefaultBind(), logger)
	ipcConfig := fmt.Sprintf(`private_key=%s
replace_peers=true
public_key=%s
endpoint=%s
replace_allowed_ips=true
allowed_ip=0.0.0.0/0
allowed_ip=::/0
persistent_keepalive_interval=%d
`, privKey, peerKey, endpoint, cfg.KeepAlive)

	if err := dev.IpcSet(ipcConfig); err != nil {
		dev.Close()
		return nil, nil, err
	}
	if err := dev.Up(); err != nil {
		dev.Close()
		return nil, nil, err
	}
	return dev, tnet, nil
}
