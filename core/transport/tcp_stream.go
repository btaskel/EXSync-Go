package transport

import (
	"net"
)

const (
	// SocketSize byte
	SocketSize    = 4096
	streamLenSize = 2
	streamIDSize  = 4
)

var (
	// defaultControlStream: create Stream and delete Stream
	// beforeStreamID, afterStreamID, flag
	// format: {0-255, 0-255, 0-255, 1-255, 0-255, 0-255, 0-255, 1-255, 0/1 (flag) }
	defaultControlStream = []byte{0, 0, 0, 1}
)

type TCPStream struct {
	*tcpWithCipher
	id              string
	idByte, comByte []byte
	rn, wn          int
	DataStart       int
}

// newTCPStream with Cipher
func newTCPStream(twc *tcpWithCipher, streamID string) (Stream, error) {
	if streamID == "" {

	}

	tc := &TCPStream{tcpWithCipher: twc,
		DataStart: twc.cipher.Info.GetIvLen() + streamIDSize - 1 + 2, // +2 DataLen
		id:        streamID,
		idByte:    []byte(streamID),
		comByte:   make([]byte, twc.socketDataLen),
	}

	err := twc.mc.CreateRecv(streamID)
	if err != nil {
		return nil, err
	}

	return tc, nil
}

// Read Stream With Cipher
// dataLen -> cipherText -> uncompressedText -> multiplexing -> plainText
func (c *TCPStream) Read(b []byte) (int, error) {
	c.rn, c.err = c.mc.PullTimeout(c.id, &b)
	if c.err != nil {
		return 0, c.err
	}
	return 0, nil
}

// Write Stream With Cipher
func (c *TCPStream) Write(b []byte) (int, error) {
	// 增加长度首部
	b[0] = byte((len(b) >> 8) & 255)
	b[1] = byte(len(b) & 255)

	// 增加流id
	copy(b[2:streamIDSize-1], c.idByte)

	// 压缩
	if c.compressor != nil {
		// dataLen +
		var n int
		n, c.err = c.compressor.CompressData(b[2+streamIDSize+1-1:], c.comByte)
		if c.err != nil || n == 0 {
			return 0, c.err
		}
	}

	// 进行加密
	if c.cipher != nil {
		c.err = c.cipher.Encrypt(c.comByte)
		if c.err != nil {
			return 0, c.err
		}
	}

	c.wn, c.err = c.conn.Write(c.comByte)
	if c.err != nil {
		return c.wn, c.err
	}
	return c.wn, c.err
}

func (c *TCPStream) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *TCPStream) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// Close 关闭一个多路复用数据流
func (c *TCPStream) Close() error {
	c.mc.Del(c.id)
	return nil
}

func (c *TCPStream) GetIVLen() int {
	return c.cipher.Info.GetIvLen()
}

func (c *TCPStream) GetKeyLen() int {
	return c.cipher.Info.GetKeyLen()
}

// NewDataSlice 创建一个数据收发切片，并返回数据起始位置
func (c *TCPStream) NewDataSlice() ([]byte, int) {
	slice := make([]byte, SocketSize)
	return slice, c.socketDataLen - 1
}
