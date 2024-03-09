package protocol

import (
	"net"
)

type NetConn interface {
	CreateStream() Reader
	ReadStream() ([]byte, error)
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	Close() error
}
