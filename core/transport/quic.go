package transport

import (
	"context"
	"crypto/tls"
	"github.com/quic-go/quic-go"
	"net"
)

// QUICDial Client Dial
func QUICDial(ctx context.Context, c net.PacketConn, addr net.Addr, tlsConf *tls.Config, conf *quic.Config) (Conn, error) {
	conn, err := quic.Dial(ctx, c, addr, tlsConf, conf)
	if err != nil {
		return nil, err
	}
	return newQUICConn(conn), nil
}

// QUICListen Server Listener
func QUICListen(addr string, tlsConf *tls.Config, config *quic.Config) (Listener, error) {
	l, err := quic.ListenAddr(addr, tlsConf, config)
	if err != nil {
		return nil, err
	}
	return &QUICListener{l}, nil
}

type QUICListener struct {
	*quic.Listener
}

func (c *QUICListener) Accept(ctx context.Context) (Conn, error) {
	conn, err := c.Listener.Accept(ctx)
	if err != nil {
		return nil, err
	}
	return newQUICConn(conn), nil
}

type QUICConn struct {
	quic.Connection
}

func newQUICConn(conn quic.Connection) *QUICConn {
	qc := QUICConn{conn}
	return &qc
}
func (c *QUICConn) OpenStream(ctx context.Context) (Stream, error) {
	stream, err := c.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	return newQUICStream(stream, c.LocalAddr(), c.RemoteAddr()), nil
}

func (c *QUICConn) Close() error {
	return nil
}
