package transport

import (
	"context"
	"github.com/quic-go/quic-go"
	"net"
)

func newTCPListener(listener net.Listener, aeadMethod, aeadPassword, compressorMethod string, tlsEnable bool) *TCPListener {
	return &TCPListener{listener, aeadMethod, aeadPassword, compressorMethod, tlsEnable}
}

type TCPListener struct {
	net.Listener
	aeadMethod       string
	aeadPassword     string
	compressorMethod string
	tlsEnable        bool
}

func (c *TCPListener) Accept(ctx context.Context) (quic.Connection, error) {
	conn, err := c.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return newTCPConn(ctx, conn, tcpConnOption{
		AEADMethod:   c.aeadMethod,
		AEADPassword: c.aeadPassword,
		Compressor:   c.compressorMethod,
		TLSEnable:    c.tlsEnable,
	})
}
