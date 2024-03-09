package protocol

import "github.com/quic-go/quic-go"

type QUICConn struct {
	session quic.Connection
}

func (q *QUICConn) CreateStream() {

}
