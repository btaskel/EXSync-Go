package transport

import (
	"EXSync/core/transport/compress"
	"EXSync/core/transport/encrypt"
	logger "EXSync/core/transport/logging"
	M "EXSync/core/transport/muxbuf"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/quic-go/quic-go"
	"net"
	"runtime"
	"sync"
)

type TCPConn struct {
	// obj
	conn       *net.TCPConn
	compressor compress.Compress
	cipher     *encrypt.Cipher
	mc         *MapChannel
	muxWriter  *muxWriter

	ctx context.Context

	taskMap           map[M.Mark]struct{} // streamID set
	streamOffsetMutex sync.Mutex
	streamOffset      uint64
	rejectNum         uint64 // stream reject counter
	maxRejectNum      uint64 // Maximum allowed rejected connections per minute

	// attr
	socketDataLen  int
	compressorLoss int
	dataGramStream quic.Stream

	applicationError *quic.ApplicationError
	errCode          quic.ApplicationErrorCode
	errMutex         sync.Mutex

	connectionState func() tls.ConnectionState
}

type tcpConnOption struct {
	AEADMethod   string
	AEADPassword string
	Compressor   string
	TLSEnable    bool
}

// newTCPConn 创建一个新的 TCPConn 实例
// ctx 是一个 context.Context，用于控制连接的生命周期
// conn 是一个 net.Conn，表示底层的网络连接
// option 是一个 tcpConnOption 结构体，包含加密和压缩选项
// 返回一个指向 TCPConn 的指针和可能的错误
func newTCPConn(ctx context.Context, conn net.Conn, option tcpConnOption) (*TCPConn, error) {
	var cipher *encrypt.Cipher
	var err error
	if option.AEADMethod != "" {
		cipher, err = encrypt.NewCipher(option.AEADMethod, option.AEADPassword)
		if err != nil {
			return nil, err
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

	tcpConn := &TCPConn{
		conn:       conn.(*net.TCPConn),
		ctx:        ctx,
		mc:         newTimeChannel(socketDataLen),
		cipher:     cipher,
		compressor: compressor,

		streamOffsetMutex: sync.Mutex{},
		streamOffset:      2, // 流ID起始位置

		socketDataLen:  socketDataLen,
		compressorLoss: n,
		maxRejectNum:   32,

		errMutex: sync.Mutex{},
	}

	tcpConn.pre()

	return tcpConn, nil
}

func (s *TCPConn) pre() {
	// 启动muxWriter
	s.muxWriter = newMuxWriter(runtime.NumCPU(), s)

	err := s.mc.createRecv(defaultControlStream)
	if err != nil {
		logger.Fatalf("connect-newTCPConn-createRecv: %s", err)
		return
	}

	go s.tcpStreamRecv()
}

func (s *TCPConn) AcceptUniStream(ctx context.Context) (quic.ReceiveStream, error) {
	//TODO implement me
	panic("implement me")
}

func (s *TCPConn) OpenUniStream() (quic.SendStream, error) {
	//TODO implement me
	panic("implement me")
}

func (s *TCPConn) OpenUniStreamSync(ctx context.Context) (quic.SendStream, error) {
	//TODO implement me
	panic("implement me")
}

// SendDatagram 与 quic.Stream 的 SendDatagram 方法实现不同,
// 它是可靠的传输, 且受AEAD与压缩、或TLS影响, 本质上是一个 Stream
func (s *TCPConn) SendDatagram(payload []byte) error {
	var err error
	if err = s.supportsDatagrams(); err != nil {
		return err
	}
	_, err = s.dataGramStream.Write(payload)
	return err
}

func (s *TCPConn) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	var n int
	var err error
	if err = s.supportsDatagrams(); err != nil {
		return nil, err
	}
	buf := make([]byte, socketLen)

	n, err = s.dataGramStream.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (s *TCPConn) supportsDatagrams() error {
	if s.dataGramStream == nil {
		return errors.New("datagram support disabled")
	}
	return nil
}

func (s *TCPConn) Context() context.Context {
	return s.ctx
}

func (s *TCPConn) ConnectionState() quic.ConnectionState {
	cs := quic.ConnectionState{}
	if s.connectionState != nil {
		cs.TLS = s.connectionState()
	}
	return cs
}

func (s *TCPConn) AcceptStream(ctx context.Context) (quic.Stream, error) {
	return s.getControlStreamByte(ctx, make([]byte, streamControlLength))
}

func (s *TCPConn) OpenStreamSync(ctx context.Context) (quic.Stream, error) {
	return s.getStream(ctx, -1)
}

func (s *TCPConn) OpenStream() (quic.Stream, error) {
	return s.getStream(s.ctx, 5)
}

func (s *TCPConn) RemoteAddr() net.Addr { return s.conn.RemoteAddr() }

func (s *TCPConn) LocalAddr() net.Addr { return s.conn.LocalAddr() }

func (s *TCPConn) CloseWithError(code quic.ApplicationErrorCode, desc string) error {
	err := s.closeLocal(code, desc)
	if err != nil {
		return err
	}

	err = s.closeRemote(code, desc)
	if err != nil {
		return err
	}

	err = s.conn.Close()
	if err != nil {
		return err
	}

	//<-s.ctx.Done()
	return nil
}

// closeLocal 携带错误代码和描述, 关闭本地套接字
func (s *TCPConn) closeLocal(code quic.ApplicationErrorCode, desc string) error {
	s.errMutex.Lock()
	defer s.errMutex.Unlock()
	if s.applicationError != nil {
		return errors.New(s.applicationError.ErrorMessage)
	}
	logger.Infof("closeLocal-host address: %s", s.conn.LocalAddr().String())
	s.applicationError = new(quic.ApplicationError)
	s.applicationError.ErrorCode = code
	s.applicationError.ErrorMessage = desc
	s.muxWriter.Close()
	s.mc.close()
	return nil
}

// closeRemote 携带错误代码和描述, 关闭远程套接字
func (s *TCPConn) closeRemote(code quic.ApplicationErrorCode, desc string) error {
	if code == 0 {
		logger.Info("Closing connection.")
	} else {
		logger.Errorf("Closing connection with error: %s", desc)
	}
	logger.Infof("Peer closed connection with error: %s", desc)

	scp, n := createStreamControlProto(streamControlProtocol{
		TypeFlag:       scpTypeFlagConn,
		BeforeStreamID: 0,
		AfterStreamID:  M.Mark(code),
		AckFlag:        0,
		ExtStr:         desc,
	})
	var err error
	_, err = s.writeStream(scp[:n], defaultControlStream)
	return err
}

// writeStream 将切片p写入到指定的流ID中
// p 是一个字节切片，表示要写入的数据
// mark 是一个 M.Mark 类型，表示目标流ID
// 返回写入的字节数和可能的错误
func (s *TCPConn) writeStream(p []byte, mark M.Mark) (int, error) {
	logger.Debug("connect-writeStream-p: ", p)
	fmt.Println("writer num", s.muxWriter.check())
	n, err := s.muxWriter.write(mark, p)
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Err == net.ErrClosed {
			if s.applicationError != nil {
				return 0, errors.New(s.applicationError.Error())
			}
		}
		return 0, err
	}
	return n, nil
}
