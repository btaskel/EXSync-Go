package transport

import (
	"net"
)

type Conn interface {
	net.Conn
	OpenStream() (Stream, error)
	CloseStream() error
}

// Stream with Cipher
type Stream interface {
	info
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close()
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

type info interface {
	GetIVLen() int
	GetKeyLen() int
}
