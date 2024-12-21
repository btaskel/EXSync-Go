package transport

import (
	"context"
	"crypto/tls"
	"github.com/quic-go/quic-go"
	"net"
	"strings"
)

// ConfOption
// 启用 TCP over TLS 时, ConfTLS 不能为 nil,
// 启用 QUIC 时, ConfQUIC 与 ConfTLS 不能为 nil,
// 启用 TCP over AEAD 时, AEADMethod AEADPassword 不能为空
type ConfOption struct {
	// ConfTLS TLS配置, 无论是否使用AEAD都将进行TLS加密,
	// 因此理想情况下应该在TLS启用时, 不使用AEAD
	ConfTLS *tls.Config
	// ConfQUIC QUIC配置
	ConfQUIC *quic.Config
	// AEADMethod AEAD加密方式,
	// 例如 encrypt.Aes128Gcm, encrypt.Xchacha20IetfPoly1305 ...
	AEADMethod string
	// AEADPassword AEAD 密钥,
	// 不限制密钥长度, 最终会哈希为加密方式所需要的密钥长度
	AEADPassword string
	// CompressorMethod 压缩方式, 例如: compress.Lz4
	CompressorMethod string
}

// DialAddr Client DialAddr
func DialAddr(ctx context.Context, network, addr string, conf *ConfOption) (quic.Connection, error) {
	switch strings.ToLower(network) {
	case "tcp":
		if conf.ConfTLS != nil {
			conn, err := tls.Dial(network, addr, conf.ConfTLS)
			if err != nil {
				return nil, err
			}
			var twc *TCPConn
			twc, err = newTCPConn(ctx, conn, tcpConnOption{
				Compressor: conf.CompressorMethod,
				TLSEnable:  true,
			})
			twc.connectionState = conn.ConnectionState
			return twc, err
		} else {
			conn, err := net.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			var twc *TCPConn
			twc, err = newTCPConn(ctx, conn, tcpConnOption{
				AEADMethod:   conf.AEADMethod,
				AEADPassword: conf.AEADPassword,
				Compressor:   conf.CompressorMethod,
			})
			if err != nil {
				return nil, err
			}

			return twc, nil
		}
	case "quic":
		conn, err := quic.DialAddr(ctx, addr, conf.ConfTLS, conf.ConfQUIC)
		if err != nil {
			return nil, err
		}
		return newQUICConn(conn), nil

	default:
		return nil, net.UnknownNetworkError(network)
	}
}

// DialTCP 它接受一个 net.Conn 并将其包装为 quic.Connection 接口
func DialTCP(ctx context.Context, c net.Conn, conf *ConfOption) (quic.Connection, error) {
	if conf.ConfTLS != nil {
		conn := tls.Client(c, conf.ConfTLS)
		err := conn.HandshakeContext(ctx)
		if err != nil {
			return nil, err
		}
		var twc *TCPConn
		twc, err = newTCPConn(ctx, conn, tcpConnOption{
			Compressor: conf.CompressorMethod,
			TLSEnable:  true,
		})
		return twc, err
	} else {
		twc, err := newTCPConn(ctx, c, tcpConnOption{
			AEADMethod:   conf.AEADMethod,
			AEADPassword: conf.AEADPassword,
			Compressor:   conf.CompressorMethod,
		})
		if err != nil {
			return nil, err
		}
		return twc, nil
	}
}

// Listen TCP, TCP over TLS, QUIC Listener
// If using TCP and tlsConf is not nil, TCP over TLS will be used
// TlsConf must exist when using QUIC
func Listen(network, addr string, Conf *ConfOption) (Listener, error) {
	switch strings.ToLower(network) {
	case "tcp":
		if Conf.ConfTLS != nil {
			listener, err := tls.Listen(network, addr, Conf.ConfTLS)
			if err != nil {
				return nil, err
			}
			return newTCPListener(listener, Conf.AEADMethod, Conf.AEADPassword, Conf.CompressorMethod, true), nil
		} else {
			listener, err := net.Listen(network, addr)
			if err != nil {
				return nil, err
			}
			return newTCPListener(listener, Conf.AEADMethod, Conf.AEADPassword, Conf.CompressorMethod, false), nil
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
