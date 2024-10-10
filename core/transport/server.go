package transport

import (
	"crypto/tls"
	"github.com/quic-go/quic-go"
	"net"
	"strings"
)

type ConfOption struct {
	ConfTLS  *tls.Config
	ConfQUIC *quic.Config

	AEADMethod       string
	AEADPassword     string
	CompressorMethod string
}

// Listen TCP, TCP over TLS, QUIC Listener
// If using TCP and tlsConf is not nil, TCP over TLS will be used
// TlsConf must exist when using QUIC
func Listen(network, addr string, Conf ConfOption) (Listener, error) {
	switch strings.ToLower(network) {
	case "tcp":
		if Conf.ConfTLS != nil {
			listener, err := tls.Listen(network, addr, Conf.ConfTLS)
			if err != nil {
				return nil, err
			}
			return newTCPListener(listener, Conf.AEADMethod, Conf.AEADPassword, Conf.CompressorMethod), nil
		} else {
			listener, err := net.Listen(network, addr)
			if err != nil {
				return nil, err
			}
			return newTCPListener(listener, Conf.AEADMethod, Conf.AEADPassword, Conf.CompressorMethod), nil
		}
	case "quic":
		l, err := quic.ListenAddr(addr, Conf.ConfTLS, Conf.ConfQUIC)
		if err != nil {
			return nil, err
		}
		return newQUICListener(l), nil
	default:
		return nil, net.UnknownNetworkError(network)
	}
}
