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

func newTCPListener(listener net.Listener, aeadMethod, aeadPassword, compressorMethod string, tlsEnable bool) *TCPListener {
	return &TCPListener{listener, aeadMethod, aeadPassword, compressorMethod, tlsEnable}
}

type tcpWithCipher struct {
	// obj
	conn       net.Conn
	tlsEnable  bool
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
	cipherLoss := getCipherLoss(c.cipher)        // 加密损耗
	compressorLoss := c.compressorLoss           // 压缩损耗
	return make([]byte, socketLen), transportLoss + cipherLoss + compressorLoss
}

type tcpConnOption struct {
	AEADMethod   string
	AEADPassword string
	Compressor   string
	TLSEnable    bool
}

func NewTCPConn(conn net.Conn, option tcpConnOption) (tcpConn *TCPConn, err error) {
	var cipher *encrypt.Cipher
	if option.AEADMethod != "" {
		cipher, err = encrypt.NewCipher(option.AEADMethod, option.AEADPassword)
		if err != nil {
			return
		}
	}

	var n int
	var compressor compress.Compress
	if option.Compressor != "" {
		compressor, n, err = compress.NewCompress(option.Compressor, socketLen-streamDataLen-streamIDLen)
		if err != nil {
			return nil, err
		}
	}

	socketDataLen := socketLen - (n + getCipherLoss(cipher) + streamIDLen)

	tcpConn = new(TCPConn)

	tcpConn.tcpWithCipher = &tcpWithCipher{
		conn:       conn,
		tlsEnable:  option.TLSEnable,
		mc:         NewTimeChannel(socketDataLen),
		cipher:     cipher,
		compressor: compressor,

		streamOffsetMutex: sync.Mutex{},
		streamOffset:      2, // 流ID起始位置

		socketDataLen:  socketDataLen,
		compressorLoss: n,
	}

	tcpConn.writeBuf = make([]byte, socketLen)

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
	writeBuf []byte
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

func (c *TCPConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *TCPConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
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

// Write
func (c *TCPConn) Write(b []byte) (int, error) {
	if len(b) <= streamDataLen+getCipherLoss(c.cipher)+streamIDLen {
		return 0, io.ErrShortBuffer
	}

	compressorWriteIndex := streamDataLen + getCipherLoss(c.cipher)
	//cipherWriteIndex := streamDataLen
	fmt.Println("tcp-write-origin:", b)
	var n int // 切片终点索引
	var srcp, dstp *[]byte
	if c.compressor == nil {
		n = len(b)
		srcp = &b
		dstp = &c.writeBuf
	} else {
		fmt.Println("pre-CompressData: ", b[compressorWriteIndex+c.compressorLoss:])
		n, c.err = c.compressor.CompressData(b[compressorWriteIndex+c.compressorLoss:], c.writeBuf[compressorWriteIndex:])
		if c.err != nil {
			return 0, c.err
		}
		fmt.Println("tcp-written", c.writeBuf)
		if n == 0 {
			// 没有压缩成功, 则使用压缩前的索引
			n = len(b)
			srcp = &c.writeBuf
			dstp = &b
		}
		n = n + compressorWriteIndex
		srcp = &c.writeBuf
		dstp = &b
	}
	fmt.Println("Compressed:", *srcp)
	if c.cipher != nil {
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

// writeStream 将切片b写入到指定的流ID中
func (c *TCPConn) writeStream(b []byte, mark M.Mark) (int, error) {
	fmt.Println("b[streamDataLen+c.cipher.Info.GetLossLen()+c.compressorLoss:]", b)
	M.CopyMarkToSlice(b[streamDataLen+getCipherLoss(c.cipher)+c.compressorLoss:], mark)
	return c.Write(b)
}
