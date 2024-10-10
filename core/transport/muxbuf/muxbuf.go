package muxbuf

import (
	"errors"
	"io"
	"time"
)

const (
	timeout = time.Duration(5) * time.Second
)

var (
	ErrTimeout = errors.New("muxbuf timeout")
)

// MuxBuf 单线程缓冲区复用结构体
type MuxBuf struct {
	freeBufChan chan *[]byte
	usedBufChan chan *[]byte

	bufNum int

	err         error
	errFlagChan chan int

	pickTimer *time.Timer
	pushTimer *time.Timer
}

// NewMuxBuf 返回一个多路复用的缓冲区, 通过传递 bufChanLen、bufLen
// 分别确定 缓冲切片存储的通道长度、缓冲切片预分配内存大小
func NewMuxBuf(bufChanLen, bufLen int) *MuxBuf {
	bufChan := make(chan *[]byte, bufChanLen)
	for i := 0; i < bufChanLen; i++ {
		slice := make([]byte, 4090, bufLen)
		bufChan <- &slice

	}

	return &MuxBuf{
		freeBufChan: bufChan,
		usedBufChan: make(chan *[]byte, bufChanLen),
		bufNum:      0,
		err:         nil,
		errFlagChan: make(chan int, 1),
		pickTimer:   time.NewTimer(timeout),
		pushTimer:   time.NewTimer(timeout),
	}
}

func pickTimeout(bufChan chan *[]byte, timer *time.Timer) (*[]byte, error) {
	timer.Reset(timeout)
	select {
	case slice := <-bufChan:
		if slice == nil {
			return nil, io.EOF
		}
		return slice, nil
	case <-timer.C:
		return nil, ErrTimeout
	}
}

// PickUsedTimeout 获取一个携带数据的切片, 如果超时则返回io.EOF
func (c *MuxBuf) PickUsedTimeout() (*[]byte, error) {
	return pickTimeout(c.usedBufChan, c.pickTimer)
}

// PickFreeTimeout 获取一个无数据的切片, 如果超时则返回io.EOF
func (c *MuxBuf) PickFreeTimeout() (*[]byte, error) {
	return pickTimeout(c.freeBufChan, c.pickTimer)
}

func pick(bufChan chan *[]byte) (*[]byte, error) {
	p := <-bufChan
	if p != nil {
		return p, nil
	}
	return nil, io.EOF
}

// PickUsed 持续堵塞, 获取一个携带数据的切片
func (c *MuxBuf) PickUsed() (*[]byte, error) {
	return pick(c.usedBufChan)
}

// PickFree 持续堵塞, 获取一个无数据的切片
func (c *MuxBuf) PickFree() (*[]byte, error) {
	return pick(c.freeBufChan)
}

// PutUsed 持续堵塞, 放置一个有数据的切片
func (c *MuxBuf) PutUsed(slice *[]byte) {
	c.usedBufChan <- slice
}

// PutFree 持续堵塞, 放置一个无数据的切片
func (c *MuxBuf) PutFree(slice *[]byte) {
	c.freeBufChan <- slice
}

// Close 终止MuxBuf
func (c *MuxBuf) Close(err error) {
	c.errFlagChan <- 1
	c.err = err
	close(c.freeBufChan)
	close(c.usedBufChan)
	close(c.errFlagChan)
}
