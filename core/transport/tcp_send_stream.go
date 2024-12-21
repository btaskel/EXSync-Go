package transport

import (
	M "EXSync/core/transport/muxbuf"
	"context"
	"errors"
	"fmt"
	"github.com/quic-go/quic-go"
	"sync"
	"time"
)

type sendStream struct {
	tcpConn *TCPConn

	mutex sync.Mutex

	ctx       context.Context
	ctxCancel context.CancelCauseFunc

	streamID M.Mark

	cancelWriteErrorCode quic.StreamErrorCode

	writeDeadLine time.Time

	compressorBuf, swapBuf       []byte
	srcBufPointer, dstBufPointer *[]byte

	writeOnce chan struct{}
	writeChan chan struct{}

	wn int

	finishedWriting bool // set once Close() is called
	cancelWriteErr  *quic.StreamError
}

func newSendStream(streamID M.Mark, tcpConn *TCPConn) (quic.SendStream, error) {
	tc := &sendStream{
		tcpConn:       tcpConn,
		streamID:      streamID,
		compressorBuf: make([]byte, tcpConn.socketDataLen),
		writeOnce:     make(chan struct{}, 1),
		mutex:         sync.Mutex{},
	}

	ctx, cancel := context.WithCancelCause(tcpConn.ctx)
	tc.ctx = ctx
	tc.ctxCancel = cancel

	if tcpConn.cipher != nil {
		tc.swapBuf = make([]byte, tcpConn.socketDataLen)
	}
	return tc, nil
}

// Write 将字节切片 b 写入到 Stream 中
// b 是要写入的数据, 返回写入的字节数和可能的错误
func (s *sendStream) Write(b []byte) (int, error) {
	s.writeOnce <- struct{}{}
	defer func() { <-s.writeOnce }()

	if s.finishedWriting {
		return 0, fmt.Errorf("write on closed stream %d", s.StreamID())
	}

	if s.cancelWriteErrorCode != 0 {
		return 0, fmt.Errorf("stream %d canceled by local with error code %d", s.streamID, s.cancelWriteErrorCode)
	}

	if !s.writeDeadLine.IsZero() && !time.Now().Before(s.writeDeadLine) {
		return 0, errDeadline
	}

	if s.tcpConn.applicationError != nil {
		return 0, errors.New(s.tcpConn.applicationError.Error())
	}

	if len(b) == 0 {
		return 0, nil
	}

	return s.tcpConn.writeStream(b, s.streamID)
}

// Close 关闭一个 quic.Stream
// 本端调用: 对端单向写入
// 本端: Read err: (由对端 Close 决定是否 io.EOF), Write err: io.EOF
// 对端: Read err: (网络缓冲区没有数据后返回 io.EOF), Write err: 正常写入
// 对端调用: 本端单向写入
// 本端: Read err: , Write err:
// 对端: Read err: , Write err:
func (s *sendStream) Close() error {
	s.mutex.Lock()
	s.finishedWriting = true
	s.mutex.Unlock()
	err := s.tcpConn.closeStream(s.streamID)
	if err != nil {
		return err
	}
	s.ctxCancel(nil)
	return nil
}

func (s *sendStream) StreamID() quic.StreamID {
	return quic.StreamID(s.streamID)
}

func (s *sendStream) CancelWrite(code quic.StreamErrorCode) {
	s.mutex.Lock()
	s.cancelWriteErrorCode = code
	s.cancelWriteImpl(code, false)
	s.mutex.Unlock()
}

func (s *sendStream) cancelWriteImpl(errorCode quic.StreamErrorCode, remote bool) {
	s.cancelWriteErr = &quic.StreamError{StreamID: s.StreamID(), ErrorCode: errorCode, Remote: remote}
	s.ctxCancel(s.cancelWriteErr)
}

func (s *sendStream) SetWriteDeadline(t time.Time) error {
	s.mutex.Lock()
	s.writeDeadLine = t
	s.mutex.Unlock()
	return nil
}

func (s *sendStream) Context() context.Context {
	return s.ctx
}
