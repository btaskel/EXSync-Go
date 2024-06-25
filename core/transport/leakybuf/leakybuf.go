package leakybuf

import (
	"errors"
	"time"
)

var (
	ErrTimeout    = errors.New("timeout")
	ErrClosedConn = errors.New("ClosedConn")
)

var timeout = time.Duration(5) * time.Second

type LeakyBuf struct {
	bufSize int
	freeBuf chan []byte
	usedBuf chan []byte
	err     error
	errChan chan error
}

// NewLeakyBuf 创建一个通道长度是n, buf大小是bufSize的LeakyBuf
func NewLeakyBuf(n, bufSize int) *LeakyBuf {
	fc := initChan(n, bufSize)
	uc := initChan(n, bufSize)
	return &LeakyBuf{
		freeBuf: fc,
		usedBuf: uc,
		bufSize: bufSize,
		errChan: make(chan error, 1),
	}
}

func initChan(n, bufSize int) chan []byte {
	c := make(chan []byte, n)
	for i := 1; i <= n; i++ {
		c <- make([]byte, 0, bufSize)
	}
	return c
}

// PickTimeout 获取已使用的LeakyBuf，并将结果存储到slice中, 该操作会修改原切片。
func (c *LeakyBuf) PickTimeout(slice *[]byte) (int, error) {
	select {
	case buf := <-c.usedBuf:
		dataLen := len(buf)
		copy(*slice, buf)
		buf = buf[:0]
		c.freeBuf <- buf
		return dataLen, nil
	case <-time.After(timeout):
		if c.err != nil {
			return 0, c.err
		}
		return 0, ErrTimeout
	}
}

// Pick 获取已使用的LeakyBuf，并将结果存储到slice中, 该操作会修改原切片，如果获取不到则阻塞。
func (c *LeakyBuf) Pick(slice *[]byte) (int, error) {
	select {
	case buf := <-c.usedBuf:
		dataLen := len(buf)
		copy(*slice, buf)
		buf = buf[:0]
		c.freeBuf <- buf
		return dataLen, nil
	case err := <-c.errChan:
		c.errChan <- err
		return 0, err
	}
}

// Put 获取未使用的LeakyBuf，并将slice存储到FreeBuf切片中
func (c *LeakyBuf) Put(slice *[]byte) error {
	select {
	case buf := <-c.freeBuf:
		copy(buf, *slice)
		c.usedBuf <- buf
		return nil
	case <-time.After(timeout):
		if c.err != nil {
			return c.err
		}
		return ErrTimeout
	}
}

// Close 设置终止错误
func (c *LeakyBuf) Close(err error) {
	close(c.freeBuf)
	close(c.usedBuf)
	c.err = err
	c.errChan <- err
}
