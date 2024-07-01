package transport

import (
	logger "EXSync/core/log"
	"EXSync/core/transport/compress"
	"EXSync/core/transport/encrypt"
	"EXSync/core/transport/leakybuf"
	"context"
	"net"
	"sync"
)

func newTCPListener(listener net.Listener, aeadMethod, compressorMethod string) *TCPListener {
	return &TCPListener{listener, aeadMethod, compressorMethod}
}

type TCPListener struct {
	net.Listener
	aeadMethod       string
	compressorMethod string
}

func (c *TCPListener) Accept(ctx context.Context) (Conn, error) {
	conn, err := c.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return NewTCPConn(conn, tcpConnOption{
		AEADMethod: c.aeadMethod,
		Compressor: c.compressorMethod,
	})
}

type streamRequest struct {
	subRejectNum int
	prevID       []byte
}

type tcpWithCipher struct {
	// obj
	conn       net.Conn
	compressor compress.Compress
	cipher     *encrypt.Cipher
	mc         *MapChannel

	taskMap           map[string]Stream // key: streamID, value: Stream obj
	streamOffsetMutex sync.Mutex
	streamOffset      uint32
	streamMap         map[string]streamRequest
	rejectNumMutex    sync.Mutex
	rejectNum         uint32 // stream reject counter
	maxRejectNum      uint32 // Maximum allowed rejected connections per minute

	// attr
	socketDataLen  int
	compressorLoss int
	err            error
}

type tcpConnOption struct {
	AEADMethod string
	Compressor string
}

func NewTCPConn(conn net.Conn, option tcpConnOption) (tcpConn *TCPConn, err error) {
	var cip *encrypt.Cipher
	if option.AEADMethod != "" {
		cip, err = encrypt.NewCipher(option.AEADMethod, option.AEADMethod)
		if err != nil {
			return
		}
	}

	var n int
	var compressor compress.Compress
	if option.Compressor != "" {
		compressor, n, err = compress.NewCompress(option.Compressor, SocketSize-2-4)
		if err != nil {
			return nil, err
		}
	}

	socketDataLen := SocketSize - (n + cip.Info.GetIvLen() + streamIDSize)

	tcpConn = &TCPConn{
		&tcpWithCipher{
			conn:       conn,
			mc:         NewTimeChannel(),
			cipher:     cip,
			compressor: compressor,

			streamOffsetMutex: sync.Mutex{},
			streamOffset:      2, // 流ID起始位置

			socketDataLen:  socketDataLen,
			compressorLoss: n,
		},
		nil,
		make([]byte, socketDataLen),
	}
	go tcpConn.tcpStreamRecv()

	tcpConn.err = tcpConn.mc.CreateRecv(string(defaultControlStream))
	if tcpConn.err != nil {
		logger.Fatalf("Transport-TCP-initDefaultStream: %s", tcpConn.err)
		return
	}

	return tcpConn, nil
}

type TCPConn struct {
	*tcpWithCipher
	defaultStream *TCPStream
	comByte       []byte
}

func (c *TCPConn) AcceptStream(ctx context.Context) (Stream, error) {
	return c.getControlStreamByte(make([]byte, c.socketDataLen))
}

func (c *TCPConn) OpenStreamSync(ctx context.Context) (Stream, error) {
	return c.getStream(ctx, -1)
}

// OpenStream 创建多路复用数据流
func (c *TCPConn) OpenStream() (Stream, error) {
	return c.getStream(context.Background(), 5)
}

// Close 释放TCPWithCipher
func (c *TCPConn) Close() error {
	c.mc.Close(leakybuf.ErrClosedConn)
	err := c.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// write 默认写入指定的Stream
func (c *TCPConn) Write(b, streamID []byte) (int, error) {
	// 增加长度首部
	b[0] = byte((len(b) >> 8) & 255)
	b[1] = byte(len(b) & 255)

	// 增加流id
	copy(b[2:streamIDSize-1], streamID)

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
	var n int
	n, c.err = c.conn.Write(c.comByte)
	if c.err != nil {
		return n, c.err
	}
	return n, c.err
}
