package leakybuf

import (
	"errors"
	"time"
)

var timeout = time.Duration(5) * time.Second

type LeakyBuf struct {
	bufSize int
	freeBuf chan []byte
	usedBuf chan []byte
	err     error
}

// NewLeakyBuf 创建一个通道长度是n, buf大小是bufSize的LeakyBuf
func NewLeakyBuf(n, bufSize int) *LeakyBuf {
	fc := initChan(n, bufSize)
	uc := initChan(n, bufSize)
	return &LeakyBuf{
		freeBuf: fc,
		usedBuf: uc,
		bufSize: bufSize,
	}
}

func initChan(n, bufSize int) chan []byte {
	c := make(chan []byte, n)
	for i := 1; i <= n; i++ {
		c <- make([]byte, 0, bufSize)
	}
	return c
}

// Pick 获取已使用的LeakyBuf，并将结果存储到slice中, 该操作会修改原切片
func (c *LeakyBuf) Pick(slice *[]byte) (int, error) {
	select {
	case buf := <-c.usedBuf:
		copy(*slice, buf)
		buf = buf[:0]
		c.freeBuf <- buf
		return len(buf), nil
	case <-time.After(timeout):
		c.err = errors.New("timeout")
		return 0, c.err
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
		c.err = errors.New("timeout")
		return c.err
	}
}

//// PickFreeBuf 获取一个空的切片, 超时则返回nil
//func (c *LeakyBuf) PickFreeBuf(slice *[]byte) error {
//	select {
//	case buf := <-c.freeBuf:
//		copy(*slice, buf[:buf[c.bufSize-1]])
//		return nil
//	case <-time.After(timeout):
//		c.err = errors.New("timeout")
//		return c.err
//	}
//}

//// PutFreeBuf 回放一个空闲切片
//func (c *LeakyBuf) PutFreeBuf(buf []byte) error {
//	if len(buf) != c.bufSize {
//		return errors.New("invalid buffer size that's put into leaky buffer")
//	}
//	buf = buf[:0]
//	c.freeBuf <- buf
//	return nil
//}
//
//// PickUsedBuf 获取一个已经使用了的切片, 超时则返回nil
//func (c *LeakyBuf) PickUsedBuf() []byte {
//	select {
//	case buf := <-c.usedBuf:
//		return buf
//	case <-time.After(timeout):
//		return nil
//	}
//}
//
//// PutUsedBuf 回放一个已经使用的切片
//func (c *LeakyBuf) PutUsedBuf(buf []byte) error {
//	if len(buf) != c.bufSize {
//		return errors.New("invalid buffer size that's put into leaky buffer")
//	}
//	select {
//	case c.usedBuf <- buf:
//	default:
//	}
//	return nil
//}

func (c *LeakyBuf) Close() {
	close(c.freeBuf)
	close(c.usedBuf)
}
