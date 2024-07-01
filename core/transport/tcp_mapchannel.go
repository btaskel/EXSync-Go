package transport

import (
	"EXSync/core/transport/leakybuf"
	"errors"
	"sync"
)

type MapChannel struct {
	mapChan map[string]*leakybuf.LeakyBuf
	lock    sync.Mutex
	err     error
}

// NewTimeChannel 创建一个数据接收队列, 每个队列默认最大使用1MB内存
func NewTimeChannel() *MapChannel {
	return &MapChannel{
		mapChan: map[string]*leakybuf.LeakyBuf{},
	}
}

// HasKey 检查一个mark是否存在于MapChannel中
func (c *MapChannel) HasKey(mark string) (ok bool) {
	_, ok = c.mapChan[mark]
	return ok
}

// CreateRecv 创建一个数据流接收队列
func (c *MapChannel) CreateRecv(mark string) (err error) {
	if !c.HasKey(mark) && len(mark) == streamIDSize {
		c.mapChan[mark] = leakybuf.NewLeakyBuf(4, 4090)
		return nil
	}
	return errors.New("markExist")
}

// Push 在一个channel中写入值
func (c *MapChannel) Push(mark string, buf []byte) error {
	c.err = c.mapChan[mark].Put(&buf)
	if c.err != nil {
		return c.err
	}
	return nil
}

// PopTimeout 获取指定mark的首部，如果超时则返回timeout错误
func (c *MapChannel) PopTimeout(mark string, buf *[]byte) (int, error) {
	n, err := c.mapChan[mark].PickTimeout(buf)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Pop 获取指定mark的首部，如果超时则返回timeout错误
func (c *MapChannel) Pop(mark string, buf *[]byte) (int, error) {
	return c.mapChan[mark].Pick(buf)
}

// Del 释放指定mark的channel对象
func (c *MapChannel) Del(mark string) {
	c.lock.Lock()
	if channel, ok := c.mapChan[mark]; ok {
		channel.Close(leakybuf.ErrClosedConn)
		delete(c.mapChan, mark)
	}
	c.lock.Unlock()
}

func (c *MapChannel) Close(err error) {
	c.lock.Lock()
	for k, v := range c.mapChan {
		v.Close(err)
		delete(c.mapChan, k)
	}
	c.lock.Unlock()
}
