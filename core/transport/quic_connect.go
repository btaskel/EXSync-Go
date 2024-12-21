package transport

import (
	"context"
	"github.com/quic-go/quic-go"
	"net"
)

func newQUICConn(conn quic.Connection) *QUICConn {
	qc := QUICConn{conn}
	return &qc
}

type QUICConn struct {
	quic.Connection
}

func (c *QUICConn) AcceptStream(ctx context.Context) (quic.Stream, error) {
	stream, err := c.Connection.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (c *QUICConn) OpenStreamSync(ctx context.Context) (quic.Stream, error) {
	stream, err := c.Connection.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (c *QUICConn) OpenStream() (quic.Stream, error) {
	stream, err := c.Connection.OpenStream()
	if err != nil {
		return nil, err
	}
	return stream, nil
}

// LocalAddr returns the local address.
func (c *QUICConn) LocalAddr() net.Addr { return c.Connection.LocalAddr() }

// RemoteAddr returns the address of the peer.
func (c *QUICConn) RemoteAddr() net.Addr { return c.Connection.RemoteAddr() }

// CloseWithError closes the connection with an error.
// The error string will be sent to the peer.
func (c *QUICConn) CloseWithError(code quic.ApplicationErrorCode, desc string) error {
	return c.Connection.CloseWithError(code, desc)
}

// Context returns a context that is cancelled when the connection is closed.
// The cancellation cause is set to the error that caused the connection to
// close, or `context.Canceled` in case the listener is closed first.
func (c *QUICConn) Context() context.Context {
	return c.Connection.Context()
}

func (c *QUICConn) ConnectionState() quic.ConnectionState {
	return c.Connection.ConnectionState()
}
