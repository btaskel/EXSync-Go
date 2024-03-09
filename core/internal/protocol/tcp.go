package protocol

import (
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/timechannel"
	"net"
)

type TCPConn struct {
	conn     net.Conn
	timeChan *timechannel.TimeChannel
}

func (t *TCPConn) CreateStream() {

}

func (t *TCPConn) ReadStream() Reader {
	tcp := TCPReader{
		tcpTimeChan: t.timeChan,
		mark:        hashext.GetRandomStr(6),
	}
	tcp.Read()
	return tcp
}

func (t *TCPConn) LocalAddr() net.Addr {
	return t.conn.LocalAddr()
}

func (t *TCPConn) RemoteAddr() net.Addr {
	return t.conn.RemoteAddr()
}
