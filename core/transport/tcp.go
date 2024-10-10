package transport

import (
	logger "EXSync/core/log"
	"EXSync/core/transport/compress"
	"EXSync/core/transport/encrypt"
	M "EXSync/core/transport/muxbuf"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
)

func newTCPListener(listener net.Listener, aeadMethod, aeadPassword, compressorMethod string) *TCPListener {
	return &TCPListener{listener, aeadMethod, aeadPassword, compressorMethod}
}

type TCPListener struct {
	net.Listener
	aeadMethod       string
	aeadPassword     string
	compressorMethod string
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
	})
}

type tcpWithCipher struct {
	// obj
	conn       net.Conn
	compressor compress.Compress
	cipher     *encrypt.Cipher
	mc         *MapChannel

	taskMap           map[M.Mark]struct{} // streamID set
	streamOffsetMutex sync.Mutex
	streamOffset      uint32
	rejectNumMutex    sync.Mutex
	rejectNum         uint32 // stream reject counter
	maxRejectNum      uint32 // Maximum allowed rejected connections per minute

	// attr
	socketDataLen  int
	compressorLoss int
	err            error
}

// getSocketSlice 初始化一个切片，并返回数据写入索引的位置
func (c *tcpWithCipher) getSocketSlice() ([]byte, int) {
	transportLoss := streamDataLen + streamIDLen // 传输层损耗
	cipherLoss := c.cipher.Info.GetLossLen()     // 加密损耗
	compressorLoss := c.compressorLoss           // 压缩损耗
	return make([]byte, SocketLen), transportLoss + cipherLoss + compressorLoss
}

type tcpConnOption struct {
	AEADMethod   string
	AEADPassword string
	Compressor   string
}

func NewTCPConn(conn net.Conn, option tcpConnOption) (tcpConn *TCPConn, err error) {
	var cip *encrypt.Cipher
	if option.AEADMethod != "" {
		cip, err = encrypt.NewCipher(option.AEADMethod, option.AEADPassword)
		if err != nil {
			return
		}
	}

	var n int
	var compressor compress.Compress
	if option.Compressor != "" {
		compressor, n, err = compress.NewCompress(option.Compressor, SocketLen-streamDataLen-streamIDLen)
		if err != nil {
			return nil, err
		}
	}

	socketDataLen := SocketLen - (n + cip.Info.GetIvLen() + streamIDLen)

	tcpConn = &TCPConn{
		defaultStream: nil,
	}

	tcpConn.tcpWithCipher = &tcpWithCipher{
		conn:       conn,
		mc:         NewTimeChannel(),
		cipher:     cip,
		compressor: compressor,

		streamOffsetMutex: sync.Mutex{},
		streamOffset:      2, // 流ID起始位置

		socketDataLen:  socketDataLen,
		compressorLoss: n,
	}

	tcpConn.buf1 = make([]byte, SocketLen)
	tcpConn.buf2 = make([]byte, SocketLen)

	go tcpConn.tcpStreamRecv()

	tcpConn.err = tcpConn.mc.CreateRecv(defaultControlStream)
	if tcpConn.err != nil {
		logger.Fatalf("Transport-TCP-initDefaultStream: %s", tcpConn.err)
		return
	}

	return tcpConn, nil
}

type TCPConn struct {
	*tcpWithCipher
	defaultStream *TCPStream
	buf1          []byte
	buf2          []byte
}

func (c *TCPConn) AcceptStream(ctx context.Context) (Stream, error) {
	_ = ctx
	return c.getControlStreamByte(make([]byte, c.socketDataLen))
}

func (c *TCPConn) OpenStreamSync(ctx context.Context) (Stream, error) {
	return c.getStream(ctx, -1)
}

func (c *TCPConn) OpenStream() (Stream, error) {
	return c.getStream(context.Background(), 5)
}

// Close 释放TCPWithCipher
func (c *TCPConn) Close() error {
	c.mc.Close(net.ErrClosed)
	err := c.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// Write 写入到指定的StreamID并发送到远程
//func (c *TCPConn) Write(buf []byte, streamID M.Mark) (int, error) {
//	// 将streamID写入到压缩与加密之前, streamID需要安全
//	index := streamDataLen + c.cipher.Info.GetLossLen() + c.compressorLoss + streamIDLen
//	fmt.Println("c.buf2[index:]", index, buf, c.buf2)
//	copy(c.buf2[index:], buf)
//	muxbuf.CopyMarkToSlice(c.buf2[index-streamIDLen:index], streamID)
//	fmt.Println("pre-push:", buf)
//	fmt.Println("pre-push-streamID:", streamID)
//	fmt.Println("pre-push-buf2:", c.buf2)
//	// 交换指针
//	srcBufPointer := &c.buf2
//	dstBufPointer := &c.buf1
//
//	// 压缩
//	var n int
//	if c.compressor != nil {
//		fmt.Println("index+len(buf)", index+len(buf))
//		fmt.Println("2+c.cipher.Info.GetLossLen()", 2+c.cipher.Info.GetLossLen())
//		n, c.err = c.compressor.CompressData(c.buf2[streamDataLen+c.cipher.Info.GetLossLen():index+len(buf)],
//			c.buf1[streamDataLen+c.cipher.Info.GetLossLen():])
//		if c.err != nil || n == 0 {
//			return 0, c.err
//		}
//		srcBufPointer = &c.buf1
//		dstBufPointer = &c.buf2
//		n = n + c.cipher.Info.GetLossLen()
//	} else {
//		n = index + len(buf)
//	}
//	fmt.Println("tcp-encry:", (*srcBufPointer)[:n])
//	// 加密
//	if c.cipher != nil {
//		c.err = c.cipher.Encrypt((*srcBufPointer)[:n], *dstBufPointer)
//		if c.err != nil {
//			return 0, c.err
//		}
//	}
//
//	// 增加长度首部
//	(*dstBufPointer)[0] = byte(((len(buf) + index) >> 8) & 255)
//	(*dstBufPointer)[1] = byte((len(buf) + index) & 255)
//
//	fmt.Println("dstBufPointer: ", *dstBufPointer)
//	fmt.Println("dstBufPointer: ", (*dstBufPointer)[:len(buf)+index+2])
//	n, c.err = c.conn.Write((*dstBufPointer)[:len(buf)+index+2])
//	if c.err != nil {
//		return n, c.err
//	}
//
//	return n, c.err
//}

func (c *TCPConn) Write(b []byte) (int, error) {
	if len(b) <= streamDataLen+c.cipher.Info.GetLossLen()+streamIDLen {
		return 0, io.ErrShortBuffer
	}

	compressorWriteIndex := streamDataLen + c.cipher.Info.GetLossLen()
	//cipherWriteIndex := streamDataLen
	fmt.Println("tcp-write-origin:", b)
	var n int // 切片终点索引
	var srcp, dstp *[]byte
	if c.compressor == nil {
		n = len(b)
		srcp = &b
		dstp = &c.buf1
	} else {
		fmt.Println("pre-CompressData: ", b[compressorWriteIndex+c.compressorLoss:])
		n, c.err = c.compressor.CompressData(b[compressorWriteIndex+c.compressorLoss:], c.buf1[compressorWriteIndex:])
		if c.err != nil {
			return 0, c.err
		}
		fmt.Println("tcp-written", c.buf1)
		if n == 0 {
			// 没有压缩成功, 则使用压缩前的索引
			n = len(b)
			srcp = &c.buf1
			dstp = &b
		}
		n = n + compressorWriteIndex
		srcp = &c.buf1
		dstp = &b
	}
	fmt.Println("Compressed:", *srcp)
	if c.cipher == nil {

	} else {
		c.err = c.cipher.Encrypt((*srcp)[:n], *dstp)
		if c.err != nil {
			return 0, c.err
		}
		srcp = dstp
	}

	(*srcp)[0] = byte(((n) >> 8) & 255)
	(*srcp)[1] = byte((n) & 255)
	fmt.Println("tcp-write-all:", *srcp)
	fmt.Println("tcp-write:", (*srcp)[:n])
	n, c.err = c.conn.Write((*srcp)[:n])
	if c.err != nil {
		return 0, c.err
	}
	return n, nil
}

func (c *TCPConn) writeStream(b []byte, mark M.Mark) (int, error) {
	fmt.Println("b[streamDataLen+c.cipher.Info.GetLossLen()+c.compressorLoss:]", b)
	M.CopyMarkToSlice(b[streamDataLen+c.cipher.Info.GetLossLen()+c.compressorLoss:], mark)
	return c.Write(b)
}
