package transport

import (
	M "EXSync/core/transport/muxbuf"
	"errors"
	"github.com/quic-go/quic-go"
	"net"
	"os"
	"time"
)

type deadlineError struct{}

func (deadlineError) Error() string   { return "deadline exceeded" }
func (deadlineError) Temporary() bool { return true }
func (deadlineError) Timeout() bool   { return true }
func (deadlineError) Unwrap() error   { return os.ErrDeadlineExceeded }

var errDeadline net.Error = new(deadlineError)

const (
	// socketLen 网络套接字发送缓冲区大小
	socketLen = 4096
	// streamDataLen 数据切片长度占用大小
	streamDataLen = 2
	// streamIDLen 流ID在数据切片中占用的大小
	streamIDLen = M.MarkLen
)

var (
	// ErrNonStreamControlProtocol 非控制流协议错误
	ErrNonStreamControlProtocol = errors.New("non-stream control protocol")
	// ErrStreamReject 控制流拒绝错误
	ErrStreamReject = errors.New("stream reject error")
)

func init() {
	if socketLen < 4 || socketLen > 65535-streamDataLen {
		panic("socketLen exceeding the index range")
	}
}

func newTCPStream(streamID M.Mark, tcpConn *TCPConn) (quic.Stream, error) {
	ts := new(TCPStream)
	var err error
	ts.ReceiveStream, err = newReceiveStream(streamID, tcpConn)
	if err != nil {
		return nil, err
	}
	ts.SendStream, err = newSendStream(streamID, tcpConn)
	if err != nil {
		return nil, err
	}
	return ts, nil
}

type TCPStream struct {
	quic.ReceiveStream
	quic.SendStream
}

func (s *TCPStream) StreamID() quic.StreamID {
	return s.SendStream.StreamID()
}

func (s *TCPStream) Close() error {
	return s.SendStream.Close()
}

func (s *TCPStream) SetDeadline(t time.Time) error {
	_ = s.SendStream.SetWriteDeadline(t)
	_ = s.ReceiveStream.SetReadDeadline(t)
	return nil
}
