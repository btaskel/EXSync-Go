package muxbuf

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

const (
	timeout = time.Second * 5
)

var (
	ErrTimeout = errors.New("muxbuf timeout")
)

// MuxBuf 单线程缓冲区复用结构体
type MuxBuf struct {
	freeBufChan chan *[]byte
	usedBufChan chan *[]byte

	swapBuf     []byte // 指向未完全读取临时存储的切片
	swapBufLock sync.Mutex

	errChan chan struct{}

	pickTimer *time.Timer

	pickTime time.Duration
}

// NewMuxBuf 返回一个多路复用的缓冲区, 通过传递 bufChanLen、bufLen
// 分别确定 缓冲切片存储的通道长度、缓冲切片预分配内存大小
func NewMuxBuf(bufChanLen, bufLen int) *MuxBuf {
	bufChan := make(chan *[]byte, bufChanLen)
	for i := 0; i < bufChanLen; i++ {
		slice := make([]byte, bufLen)
		bufChan <- &slice
	}

	return &MuxBuf{
		freeBufChan: bufChan,
		usedBufChan: make(chan *[]byte, bufChanLen),
		errChan:     make(chan struct{}, 1),
		swapBufLock: sync.Mutex{},
		pickTime:    timeout,
		pickTimer:   time.NewTimer(timeout),
	}
}

// SetSwapBuf 设置 swapBuf 记录
func (s *MuxBuf) SetSwapBuf(buf []byte) {
	s.swapBuf = buf
}

// GetSwapBuf 获取 swapBuf 记录
func (s *MuxBuf) GetSwapBuf() []byte { return s.swapBuf }

// GetUsedBufLen 获取UsedBuf长度
func (s *MuxBuf) GetUsedBufLen() int { return len(s.usedBufChan) }

// GetFreeBufLen 获取FreeBuf长度
func (s *MuxBuf) GetFreeBufLen() int { return len(s.usedBufChan) }

// SetPickTime 设置 pickTimeout 最大等待时间
func (s *MuxBuf) SetPickTime(t time.Duration) {
	s.pickTime = t
}

func (s *MuxBuf) pickTimeout(bufChan chan *[]byte, timer *time.Timer) (*[]byte, error) {
	timer.Reset(s.pickTime)
	select {
	case slice := <-bufChan:
		if slice != nil {
			return slice, nil
		}
		return nil, io.EOF
	case <-timer.C:
		return nil, ErrTimeout
	case <-s.errChan:
		return nil, net.ErrClosed
	}
}

// PickUsedTimeout 如果使用 SetReadDeadline 它将会启用超时
// 获取一个携带数据的切片, 如果超时则返回io.EOF
func (s *MuxBuf) PickUsedTimeout() (*[]byte, error) {
	if s.pickTime != 0 {
		return s.pickTimeout(s.usedBufChan, s.pickTimer)
	}
	return s.pick(s.usedBufChan)
}

// PickFreeTimeout 获取一个无数据的切片, 如果超时则返回io.EOF
func (s *MuxBuf) PickFreeTimeout() (*[]byte, error) {
	if s.pickTime != 0 {
		return s.pickTimeout(s.freeBufChan, s.pickTimer)
	}
	return s.pick(s.freeBufChan)
}

func (s *MuxBuf) pick(bufChan chan *[]byte) (*[]byte, error) {
	if s.swapBuf != nil {
		defer func() { s.swapBuf = nil }()
		return &s.swapBuf, nil
	}
	if p := <-bufChan; p != nil {
		return p, nil
	}
	return nil, io.EOF
}

// PickUsed 持续堵塞, 获取一个携带数据的切片
func (s *MuxBuf) PickUsed() (*[]byte, error) {
	return s.pick(s.usedBufChan)
}

// PickFree 持续堵塞, 获取一个无数据的切片
func (s *MuxBuf) PickFree() (*[]byte, error) {
	return s.pick(s.freeBufChan)
}

// PutUsed 持续堵塞, 放置一个有数据的切片
func (s *MuxBuf) PutUsed(slice *[]byte) {
	s.usedBufChan <- slice
}

// PutFree 持续堵塞, 放置一个无数据的切片
func (s *MuxBuf) PutFree(slice *[]byte) {
	s.freeBufChan <- slice
}

// Close 终止MuxBuf
func (s *MuxBuf) Close() {
	s.errChan <- struct{}{}
	close(s.freeBufChan)
	close(s.usedBufChan)
	close(s.errChan)
}
