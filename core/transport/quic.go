package transport

import (
	"context"
	"github.com/quic-go/quic-go"
)

func newQUICListener(listener *quic.Listener) *QUICListener {
	return &QUICListener{listener}
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

func newQUICConn(conn quic.Connection) *QUICConn {
	qc := QUICConn{conn}
	return &qc
}

type QUICConn struct {
	quic.Connection
}

func (c *QUICConn) AcceptStream(ctx context.Context) (Stream, error) {
	stream, err := c.Connection.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (c *QUICConn) OpenStreamSync(ctx context.Context) (Stream, error) {
	stream, err := c.Connection.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (c *QUICConn) OpenStream() (Stream, error) {
	stream, err := c.Connection.OpenStream()
	if err != nil {
		return nil, err
	}
	return newQUICStream(stream, c.LocalAddr(), c.RemoteAddr()), nil
}

func (c *QUICConn) Close() error {
	return nil
}
