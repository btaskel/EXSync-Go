package transport

import (
	"github.com/quic-go/quic-go"
	"net"
)

type QUICStream struct {
	quic.Stream
	localAddr, remoteAddr net.Addr
}

func newQUICStream(stream quic.Stream, localAddr, remoteAddr net.Addr) *QUICStream {
	return &QUICStream{
		Stream:     stream,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}
}

func (c *QUICStream) LocalAddr() net.Addr {
	return c.localAddr
}
func (c *QUICStream) RemoteAddr() net.Addr {
	return c.remoteAddr
}
