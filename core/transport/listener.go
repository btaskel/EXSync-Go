package transport

import (
	"context"
	"net"
)

type Listener interface {
	Addr() net.Addr
	Close() error
	Accept(ctx context.Context) (Conn, error)
}
