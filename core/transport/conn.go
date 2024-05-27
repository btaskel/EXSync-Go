package transport

import (
	"context"
	"net"
)

type Conn interface {
	OpenStream(ctx context.Context) (Stream, error)
	Close() error
}

// Stream with Cipher
type Stream interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}
