package transport

import (
	"context"
	"crypto/tls"
	"github.com/quic-go/quic-go"
	"net"
	"strings"
)

// Dial Client Dial
func Dial(ctx context.Context, network, addr string, Conf ConfOption) (Conn, error) {
	switch strings.ToLower(network) {
	case "tcp":
		if Conf.ConfTLS != nil {
			conn, err := tls.Dial(network, addr, Conf.ConfTLS)
			if err != nil {
				return nil, err
			}
			var twc *TCPConn
			twc, err = NewTCPConn(conn, tcpConnOption{
				Compressor: Conf.CompressorMethod,
			})
			return twc, err

		} else {
			conn, err := net.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			var twc *TCPConn
			twc, err = NewTCPConn(conn, tcpConnOption{
				AEADMethod:   Conf.AEADMethod,
				AEADPassword: Conf.AEADPassword,
				Compressor:   Conf.CompressorMethod,
			})
			return twc, nil
		}

	case "quic":
		conn, err := quic.DialAddr(ctx, addr, Conf.ConfTLS, Conf.ConfQUIC)
		if err != nil {
			return nil, err
		}
		return newQUICConn(conn), nil

	default:
		return nil, net.UnknownNetworkError(network)
	}
}
