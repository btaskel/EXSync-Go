package transport

import (
	M "EXSync/core/transport/muxbuf"
	"net"
	"sync"
)

type MapChannel struct {
	mapChan map[M.Mark]*M.MuxBuf
	lock    sync.Mutex
	err     error
	bufLen  int
}

// NewTimeChannel 创建一个数据接收队列, 每个队列默认最大使用1MB内存
func NewTimeChannel(bufLen int) *MapChannel {
	return &MapChannel{
		mapChan: make(map[M.Mark]*M.MuxBuf, 32),
		bufLen:  bufLen,
	}
}

// GetMuxBuf 获取一个MuxBuf
func (c *MapChannel) GetMuxBuf(mark M.Mark) (*M.MuxBuf, bool) {
	muxBuf, ok := c.mapChan[mark]
	return muxBuf, ok
}

// HasKey 检查一个mark是否存在于MapChannel中
func (c *MapChannel) HasKey(mark M.Mark) (ok bool) {
	_, ok = c.mapChan[mark]
	return ok
}

// CreateRecv 创建一个数据流接收队列
func (c *MapChannel) CreateRecv(mark M.Mark) (err error) {
	if !c.HasKey(mark) {
		// 创建一个有128队列的数据部分大小的网络接收缓冲区
		c.mapChan[mark] = M.NewMuxBuf(128, c.bufLen)
		return nil
	}
	return M.ErrMarkExist
}

func (c *MapChannel) push(muxBufPickFunc func() (*[]byte, error), putUsedFunc func(*[]byte), buf *[]byte) error {
	var pushPointer *[]byte
	pushPointer, c.err = muxBufPickFunc()
	if c.err != nil {
		return c.err
	}
	*pushPointer = (*pushPointer)[:len(*buf)] // 调整自由切片大小为当前写入切片大小
	copy(*pushPointer, *buf)
	putUsedFunc(pushPointer)
	return nil
}

// PushTimeout 在一个channel中写入值, 如果关闭了muxBuf则返回EOF
func (c *MapChannel) PushTimeout(muxBuf *M.MuxBuf, buf *[]byte) error {
	return c.push(muxBuf.PickFreeTimeout, muxBuf.PutUsed, buf)
}

// Push 在一个muxBuf中写入缓冲
func (c *MapChannel) Push(muxBuf *M.MuxBuf, buf *[]byte) error {
	err := c.push(muxBuf.PickFree, muxBuf.PutUsed, buf)
	if err != nil {
		return err
	}
	return nil
}

func (c *MapChannel) pop(muxBufFunc func() (*[]byte, error), buf *[]byte) (int, error) {
	var popPointer *[]byte
	popPointer, c.err = muxBufFunc()
	if c.err != nil {
		return 0, c.err
	}
	copy(*buf, *popPointer)
	return len(*popPointer), nil
}

// PopTimeout 获取指定mark的首部，如果超时则返回timeout错误, 如果关闭了muxBuf则返回EOF
func (c *MapChannel) PopTimeout(muxBuf *M.MuxBuf, buf *[]byte) (int, error) {
	return c.pop(muxBuf.PickUsedTimeout, buf)
}

// Pop 获取指定mark的首部，如果超时则返回timeout错误, 如果关闭了muxBuf则返回EOF
func (c *MapChannel) Pop(muxBuf *M.MuxBuf, buf *[]byte) (int, error) {
	return c.pop(muxBuf.PickUsed, buf)
}

// Del 释放指定mark的channel对象
func (c *MapChannel) Del(mark M.Mark) error {
	if !c.HasKey(mark) {
		return M.ErrMarkNotExist
	}
	c.lock.Lock()
	c.mapChan[mark].Close(net.ErrClosed)
	delete(c.mapChan, mark)
	c.lock.Unlock()
	return nil
}

func (c *MapChannel) Close(err error) {
	c.lock.Lock()
	for k, v := range c.mapChan {
		v.Close(err)
		delete(c.mapChan, k)
	}
	c.lock.Unlock()
}
