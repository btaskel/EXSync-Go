package transport

import (
	M "EXSync/core/transport/muxbuf"
	"context"
	"errors"
	"fmt"
	"github.com/quic-go/quic-go"
	"io"
	"net"
	"sync"
	"time"
)

type receiveStream struct {
	tcpConn *TCPConn
	muxBuf  *M.MuxBuf

	mutex sync.Mutex

	ctx       context.Context
	ctxCancel context.CancelCauseFunc

	streamID M.Mark

	cancelReadErrorCode quic.StreamErrorCode
	cancelledLocally    bool

	readDeadLine time.Time

	readOnce chan struct{}
	readChan chan struct{}

	readDeadline  time.Time
	writeDeadline time.Time

	rn int

	cancelReadErr *quic.StreamError

	errorRead         bool
	cancelledRemotely bool
}

// newReceiveStream with Cipher
// 如果streamID为空，则自动选择StreamID
func newReceiveStream(streamID M.Mark, tcpConn *TCPConn) (quic.ReceiveStream, error) {
	tc := &receiveStream{
		streamID: streamID,
		tcpConn:  tcpConn,
		muxBuf:   tcpConn.mc.mapChan[streamID],
		readOnce: make(chan struct{}, 1),
		mutex:    sync.Mutex{},
	}

	ctx, cancel := context.WithCancelCause(tcpConn.ctx)
	tc.ctx = ctx
	tc.ctxCancel = cancel

	return tc, nil
}

// Read quic.Stream With Cipher
// dataLen -> cipherText -> uncompressedText -> multiplexing -> plainText
func (s *receiveStream) Read(p []byte) (int, error) {
	s.readOnce <- struct{}{}
	defer func() { <-s.readOnce }()

	if s.cancelReadErrorCode != 0 {
		return 0, fmt.Errorf("stream %d canceled by local with error code %d", s.streamID, s.cancelReadErrorCode)
	}

	if !s.readDeadLine.IsZero() && !time.Now().Before(s.readDeadLine) {
		return 0, errDeadline
	}
	return s.readImpl(p)
}

func (s *receiveStream) readImpl(p []byte) (int, error) {
	var err error
	s.rn, err = s.tcpConn.mc.pop(s.muxBuf, &p)
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Err == net.ErrClosed || opErr.Err == io.EOF {
			if s.tcpConn.applicationError != nil {
				return 0, errors.New(s.tcpConn.applicationError.Error())
			}
		}
		return 0, err
	}
	return s.rn, nil
}

func (s *receiveStream) StreamID() quic.StreamID {
	return quic.StreamID(s.streamID)
}

func (s *receiveStream) CancelRead(code quic.StreamErrorCode) {
	s.mutex.Lock()
	s.cancelReadErrorCode = code
	s.cancelReadImpl(code)
	s.mutex.Unlock()
}

func (s *receiveStream) cancelReadImpl(errorCode quic.StreamErrorCode) bool {
	if s.cancelledLocally {
		return false
	}
	s.cancelledLocally = true
	if s.errorRead || s.cancelledRemotely {
		return false
	}
	s.cancelReadErr = &quic.StreamError{StreamID: s.StreamID(), ErrorCode: errorCode, Remote: false}
	s.ctxCancel(s.cancelReadErr)
	return true
}

func (s *receiveStream) SetReadDeadline(t time.Time) error {
	s.mutex.Lock()
	s.readDeadLine = t
	s.mutex.Unlock()
	return nil
}
