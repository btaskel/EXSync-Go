package protocol

import (
	"github.com/quic-go/quic-go"
	"net"
)

func NewNewConn(protocol string, conn any) (NetConn, error) {
	var quicConn quic.Connection
	var tcpConn net.Conn

	switch v := conn.(type) {
	case quic.Connection:
		quicConn = v
	case net.Conn:
		tcpConn = v
	default:
		panic("传入未知的网络类型")
	}

	switch protocol {
	case "tcp":
		return TCPConn{conn: tcpConn}, nil
	case "quic":
		return QUICConn{quicConn}, nil
	}
}
