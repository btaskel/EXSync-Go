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

func (c *QUICListener) Accept(ctx context.Context) (quic.Connection, error) {
	conn, err := c.Listener.Accept(ctx)
	if err != nil {
		return nil, err
	}
	return newQUICConn(conn), nil
}
