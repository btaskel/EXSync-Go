package protocol

import (
	"EXSync/core/internal/modules/timechannel"
	"github.com/quic-go/quic-go"
)

// Reader 对流的操作接口
type Reader interface {
	Read(b []byte) error
}

type TCPReader struct {
	tcpTimeChan *timechannel.TimeChannel
	mark        string
}

func (t *TCPReader) Read(b []byte) error {
	data, err := t.tcpTimeChan.Get(t.mark)
	if err != nil {
		return err
	}
	return nil
}

type QuicReader struct {
	quicStream quic.Stream
}

func (q *QuicReader) Read(b []byte) error {
	_, err := q.quicStream.Read(b)
	if err != nil {
		return err
	}
	return nil
}

//type Stream struct {
//	tcpTimeChan *timechannel.TimeChannel
//	quicStream  *quic.Stream
//}
//
//func (s *Stream) Read(b []byte, n int) {
//	if s.tcpTimeChan != nil{
//		...
//	} else {
//		...
//	}
//}
