package transport

import (
	M "EXSync/core/transport/muxbuf"
	"errors"
	"net"
)

const (
	// socketLen 网络套接字发送缓冲区大小
	socketLen = 4096
	// streamDataLen 数据切片长度占用大小
	streamDataLen = 2
	// streamIDLen 流ID在数据切片中占用的大小
	streamIDLen = 4
)

var (
	// defaultControlStream: create Stream and delete Stream
	// delete-flag, beforeStreamID, afterStreamID, reply-flag
	// format: {0/1 (delete-flag) ,0-255, 0-255, 0-255, 1-255, 0-255, 0-255, 0-255, 1-255, 0/1 (reply-flag) }
	defaultControlStream = M.Mark{0, 0, 0, 1}
)

var (
	// ErrStreamClosed 流已关闭
	ErrStreamClosed = errors.New("stream closed")
	// ErrNonStreamControlProtocol 非控制流协议错误
	ErrNonStreamControlProtocol = errors.New("non-stream control protocol")
	// ErrStreamReject 控制流拒绝错误
	ErrStreamReject = errors.New("stream reject error")
)

type TCPStream struct {
	*tcpWithCipher
	streamID M.Mark

	compressorBuf, swapBuf       []byte
	srcBufPointer, dstBufPointer *[]byte

	tcpConnection *TCPConn
	rn, wn        int
}

// newTCPStream with Cipher
// 如果streamID为空，则自动选择StreamID
func newTCPStream(twc *tcpWithCipher, streamID M.Mark, tcpConn *TCPConn) (Stream, error) {
	tc := &TCPStream{tcpWithCipher: twc,
		streamID:      streamID,
		compressorBuf: make([]byte, twc.socketDataLen),
		tcpConnection: tcpConn,
	}

	if twc.cipher != nil {
		tc.swapBuf = make([]byte, twc.socketDataLen)
	}
	return tc, nil
}

// Read Stream With Cipher
// dataLen -> cipherText -> uncompressedText -> multiplexing -> plainText
func (c *TCPStream) Read(b []byte) (int, error) {
	var err error
	c.rn, err = c.mc.PopTimeout(c.mc.mapChan[c.streamID], &b)
	if err != nil {
		return 0, err
	}
	return c.rn, nil
}

func (c *TCPStream) Write(b []byte) (int, error) {
	return c.tcpConnection.writeStream(b, c.streamID)
}

func (c *TCPStream) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *TCPStream) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// Close 关闭一个多路复用数据流
func (c *TCPStream) Close() error {
	return c.mc.Del(c.streamID)
}

func (c *TCPStream) getIVLen() int {
	return c.cipher.Info.GetIvLen()
}

func (c *TCPStream) getKeyLen() int {
	return c.cipher.Info.GetKeyLen()
}

// GetBuf 初始化一个切片，并返回数据写入索引的位置
func (c *TCPStream) GetBuf() ([]byte, int) {
	return c.getSocketSlice()
}
