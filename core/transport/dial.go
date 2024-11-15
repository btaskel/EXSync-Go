package transport

import (
	"context"
	"crypto/tls"
	"github.com/quic-go/quic-go"
	"golang.org/x/net/proxy"
	"net"
	"strings"
)

// DialAddr Client DialAddr
func DialAddr(ctx context.Context, network, addr string, conf *ConfOption) (Conn, error) {
	switch strings.ToLower(network) {
	case "tcp":
		if conf.ConfTLS != nil {
			_ = ctx
			conn, err := tls.Dial(network, addr, conf.ConfTLS)
			if err != nil {
				return nil, err
			}
			var twc *TCPConn
			twc, err = NewTCPConn(conn, tcpConnOption{
				Compressor: conf.CompressorMethod,
				TLSEnable:  true,
			})
			return twc, err
		} else {
			conn, err := net.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			var twc *TCPConn
			twc, err = NewTCPConn(conn, tcpConnOption{
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

// DialTCP 它接受一个 net.Conn 并将其包装为 transport.Conn 接口
func DialTCP(ctx context.Context, c net.Conn, conf *ConfOption) (Conn, error) {
	if conf.ConfTLS != nil {
		conn := tls.Client(c, conf.ConfTLS)
		err := conn.HandshakeContext(ctx)
		if err != nil {
			return nil, err
		}
		var twc *TCPConn
		twc, err = NewTCPConn(conn, tcpConnOption{
			Compressor: conf.CompressorMethod,
			TLSEnable:  true,
		})
		return twc, err
	} else {
		twc, err := NewTCPConn(c, tcpConnOption{
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

//func tlsWithProxyDial(ctx context.Context, network, addr string, Conf *ConfOption) (Conn, error) {
//	var conn net.Conn
//	var err error
//	var dialer proxy.Dialer
//	dialer, err = proxy.SOCKS5("tcp", Conf.ProxyAddr, Conf.ProxyAuth, proxy.Direct)
//	if err != nil {
//		return nil, err
//	}
//
//	conn, err = tls.DialWithDialer(&net.Dialer{
//		Timeout:   10 * time.Second, // 设置拨号超时时间
//		KeepAlive: 10 * time.Second,
//		Control: func(_, addr string, c syscall.RawConn) error {
//			_, err = dialer.(proxy.ContextDialer).DialContext(ctx, "tcp", addr)
//			if err != nil {
//				return err
//			}
//			return err
//		},
//	}, network, addr, Conf.ConfTLS)
//
//	if err != nil {
//		return nil, err
//	}
//	var twc *TCPConn
//	twc, err = NewTCPConn(conn, tcpConnOption{
//		Compressor: Conf.CompressorMethod,
//	})
//	return twc, err
//
//}
//
//func tlsDial(ctx context.Context, network, addr string, Conf *ConfOption) (Conn, error) {
//	_ = ctx
//	conn, err := tls.Dial(network, addr, Conf.ConfTLS)
//	if err != nil {
//		return nil, err
//	}
//	var twc *TCPConn
//	twc, err = NewTCPConn(conn, tcpConnOption{
//		Compressor: Conf.CompressorMethod,
//	})
//	return twc, err
//}

// ConfOption 优先级顺序:
type ConfOption struct {
	// ConfTLS TLS配置, 并开启TCP,无论是否开启AEAD都将使用tls加密
	ConfTLS *tls.Config
	// ConfQUIC Quic配置
	ConfQUIC *quic.Config

	AEADMethod       string
	AEADPassword     string
	CompressorMethod string

	ProxyAddr string
	ProxyAuth *proxy.Auth
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
