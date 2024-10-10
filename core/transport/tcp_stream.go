package transport

import (
	M "EXSync/core/transport/muxbuf"
	"errors"
	"net"
)

const (
	// SocketLen
	SocketLen = 4096
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

// Write Stream With Cipher
// 会复用 b 作为发送切片, 因此需要在 2 + c.cipher.Info.GetLossLen() + c.compressorLoss + streamIDLen 之后写入数据
//func (c *TCPStream) Write(b []byte) (int, error) {
//	fmt.Println("w-b", b)
//	fmt.Println("b len", len(b))
//	muxbuf.CopyMarkToSlice(b[streamDataLen+c.cipher.Info.GetLossLen()+c.compressorLoss+1:], c.streamID)
//
//	if c.compressor != nil {
//		fmt.Println("CompressData:", b[streamDataLen+c.cipher.Info.GetLossLen()+c.compressorLoss:])
//		c.wn, c.err = c.compressor.CompressData(b[streamDataLen+c.cipher.Info.GetLossLen()+c.compressorLoss:],
//			c.compressorBuf[streamDataLen+c.cipher.Info.GetLossLen():])
//		if c.err != nil {
//			return 0, c.err
//		}
//		if c.wn == 0 {
//			fmt.Println("c.wn reset!")
//			c.wn = len(b) - streamDataLen + c.cipher.Info.GetLossLen()
//		}
//
//		c.srcBufPointer = &c.compressorBuf
//		c.dstBufPointer = &c.swapBuf
//		c.wn += streamDataLen + c.cipher.Info.GetLossLen()
//	} else {
//		c.wn = len(b)
//		c.srcBufPointer = &b
//		c.dstBufPointer = &c.swapBuf
//	}
//
//	// [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 2 50 51 52 53]
//	// [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 2 0 0 0 2 1 0 0]
//	fmt.Println("wn:", c.wn, b)
//	if c.cipher != nil {
//		fmt.Println("Encrypt", (*c.srcBufPointer)[:c.wn])
//		fmt.Println("Encrypt-len", len((*c.srcBufPointer)[:c.wn]))
//		c.err = c.cipher.Encrypt((*c.srcBufPointer)[:c.wn],
//			*c.dstBufPointer)
//		fmt.Println("cipher.Encrypted-all:", *c.dstBufPointer)
//		fmt.Println("cipher.Encrypted:", (*c.dstBufPointer)[:c.wn])
//		fmt.Println("cipher.Encrypted-len:", len((*c.dstBufPointer)[:c.wn]))
//		if c.err != nil {
//			return 0, c.err
//		}
//		c.srcBufPointer = c.dstBufPointer
//	}
//
//	// 增加长度首部
//	(*c.srcBufPointer)[0] = byte(((c.wn - streamDataLen) >> 8) & 255)
//	(*c.srcBufPointer)[1] = byte((c.wn - streamDataLen) & 255)
//	fmt.Println("stream-write: ", (*c.srcBufPointer)[:c.wn])
//	fmt.Println("stream-write-len: ", len((*c.srcBufPointer)[:c.wn]))
//	c.wn, c.err = c.conn.Write((*c.srcBufPointer)[:c.wn])
//	if c.err != nil {
//		return c.wn, c.err
//	}
//	return c.wn, c.err
//}

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
