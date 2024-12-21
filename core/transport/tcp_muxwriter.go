package transport

import (
	logger "EXSync/core/transport/logging"
	"EXSync/core/transport/muxbuf"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

type muxWriterI struct {
	tcpConn *TCPConn

	dataBuf, writeBuf []byte
	srcp, dstp        *[]byte

	compressorWriteIndex int

	wn, sentData, cn int

	err error
}

func newWriter(tcpConn *TCPConn) *muxWriterI {
	compressorWriteIndex := streamDataLen
	if tcpConn.cipher != nil {
		compressorWriteIndex += tcpConn.cipher.Info.GetLossLen()
	}
	mwi := &muxWriterI{
		tcpConn:              tcpConn,
		compressorWriteIndex: compressorWriteIndex,
	}
	mwi.writeBuf = make([]byte, socketLen)
	mwi.dataBuf = make([]byte, socketLen)
	return mwi
}

func (s *muxWriterI) compress() {
	//fmt.Println("tcp-write-origin:", s.dataBuf)
	if s.tcpConn.compressor == nil {
		s.wn = len(s.dataBuf)
		s.srcp = &s.dataBuf
		s.dstp = &s.writeBuf
	} else {
		//fmt.Println("pre-CompressData: ", s.dataBuf[s.compressorWriteIndex+s.tcpConn.compressor.GetLoss():])
		s.wn, s.err = s.tcpConn.compressor.CompressData(s.dataBuf[s.compressorWriteIndex+s.tcpConn.compressor.GetLoss():], s.writeBuf[s.compressorWriteIndex:])
		if s.err != nil {
			return
		}
		//fmt.Println("tcp-written", s.writeBuf)
		if s.wn == 0 {
			// 没有压缩成功, 则使用压缩前的索引
			s.wn = len(s.dataBuf)
		} else {
			s.wn += s.compressorWriteIndex
		}
		s.srcp = &s.writeBuf
		s.dstp = &s.dataBuf
	}
}

func (s *muxWriterI) encrypt() {
	if s.tcpConn.cipher != nil {
		s.err = s.tcpConn.cipher.Encrypt((*s.srcp)[:s.wn], *s.dstp)
		if s.err != nil {
			return
		}
		s.srcp = s.dstp
	}
}

func (s *muxWriterI) setMark(mark muxbuf.Mark) {
	var cipherLoss int
	if s.tcpConn.cipher == nil {
		cipherLoss = 0
	} else {
		cipherLoss = s.tcpConn.cipher.Info.GetLossLen()
	}
	binary.BigEndian.PutUint64(s.dataBuf[streamDataLen+cipherLoss+s.tcpConn.compressor.GetLoss():], mark)
	//CopyMarkToSlice(s.dataBuf[s.streamDataLen+cipherLoss+s.compressor.GetLoss():], mark)
}

func (s *muxWriterI) write(b []byte) (int, error) {
	s.sentData = 0
	dataIndex := s.tcpConn.compressor.GetLoss() + s.compressorWriteIndex + streamIDLen
	counter := 0
	for {
		fmt.Println("counter: ", counter)
		counter++
		if counter == 5 {
			panic("counter err")
		}

		logger.Debugf("muxwriter-write-s.dataBuf: %v", s.dataBuf)

		s.cn = copy(s.dataBuf[dataIndex:], b[s.sentData:])
		s.dataBuf = s.dataBuf[:dataIndex+s.cn]
		s.compress()
		if s.err != nil {
			return 0, s.err
		}
		logger.Debugf("muxwriter-write-Compressed: %v", *s.srcp)
		s.encrypt()
		if s.err != nil {
			return 0, s.err
		}
		(*s.srcp)[0] = byte(((s.wn) >> 8) & 255)
		(*s.srcp)[1] = byte((s.wn) & 255)
		logger.Debugf("muxwriter-write-sent-b: %v", b)
		logger.Infof("muxwriter-write-sent-Data: %v", (*s.srcp)[:s.wn])
		s.wn, s.err = s.tcpConn.conn.Write((*s.srcp)[:s.wn])
		if s.err != nil {
			return 0, s.err
		}
		s.sentData += s.cn
		logger.Infof("muxwriter-write-sent-len: %v", s.sentData)
		if s.sentData == len(b) {
			return s.sentData, nil
		}
	}
}

type muxWriter struct {
	writerChan chan *muxWriterI
	closeFlag  bool
	errMutex   sync.Mutex
	closeOnce  sync.Once
}

// newMuxWriter 创建一个Writer池, 以便避免在并发传输时减少内存重分配或单线程时锁带来的额外开销
func newMuxWriter(size int, tcpConn *TCPConn) *muxWriter {
	mw := new(muxWriter)
	mw.errMutex = sync.Mutex{}
	mw.closeOnce = sync.Once{}
	mw.writerChan = make(chan *muxWriterI, size)
	for i := 0; i < size; i++ {
		mw.writerChan <- newWriter(tcpConn)
	}
	return mw
}

func (s *muxWriter) write(mark muxbuf.Mark, p []byte) (int, error) {
	mwi := <-s.writerChan
	if mwi == nil {
		return 0, net.ErrClosed
	}
	defer func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("muxWriter-write-recover: %v", r)
			}
		}()
		if !s.closeFlag {
			select {
			case s.writerChan <- mwi:
			default:
			}
		}
	}()
	mwi.setMark(mark)
	return mwi.write(p)
}

func (s *muxWriter) check() int {
	return len(s.writerChan)
}

// Close 关闭 muxWriterI 通道
// Write 会因此返回 net.ErrClosed
func (s *muxWriter) Close() {
	s.closeOnce.Do(
		func() {
			s.closeFlag = true
			close(s.writerChan)
		})
}
