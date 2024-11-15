package transport

import (
	"context"
	"net"
)

type TCPListener struct {
	net.Listener
	aeadMethod       string
	aeadPassword     string
	compressorMethod string
	tlsEnable        bool
}

func (c *TCPListener) Accept(ctx context.Context) (Conn, error) {
	_ = ctx
	conn, err := c.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return NewTCPConn(conn, tcpConnOption{
		AEADMethod:   c.aeadMethod,
		AEADPassword: c.aeadPassword,
		Compressor:   c.compressorMethod,
		TLSEnable:    c.tlsEnable,
	})
}
