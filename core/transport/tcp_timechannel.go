package transport

import (
	"EXSync/core/transport/leakybuf"
	"errors"
	"sync"
)

type MapChannel struct {
	channelDict map[string]*leakybuf.LeakyBuf
	lock        sync.Mutex
	err         error
}

// NewTimeChannel 创建一个数据接收队列, 每个队列默认最大使用1MB内存
func NewTimeChannel() *MapChannel {
	return &MapChannel{
		channelDict: map[string]*leakybuf.LeakyBuf{},
	}
}

// HasKey 检查一个mark是否存在于MapChannel中
func (c *MapChannel) HasKey(mark string) (ok bool) {
	_, ok = c.channelDict[mark]
	return ok
}

// CreateRecv 创建一个数据流接收队列
func (c *MapChannel) CreateRecv(mark string) (err error) {
	if !c.HasKey(mark) && len(mark) == streamIDSize {
		c.channelDict[mark] = leakybuf.NewLeakyBuf(4, 4090)
		return nil
	}
	return errors.New("markExist")
}

// Push 在一个channel中写入值
func (c *MapChannel) Push(mark string, buf []byte) error {
	c.err = c.channelDict[mark].Put(&buf)
	if c.err != nil {
		return c.err
	}
	return nil
}

// Pull 获取指定mark的首部，如果超时则返回timeout错误
func (c *MapChannel) Pull(mark string, buf *[]byte) (int, error) {
	n, err := c.channelDict[mark].Pick(buf)
	if err != nil {
		return n, err
	}
	return n, nil
}

// Del 释放指定mark的channel对象
func (c *MapChannel) Del(mark string) {
	c.lock.Lock()
	if channel, ok := c.channelDict[mark]; ok {
		channel.Close()
		delete(c.channelDict, mark)
	}
	c.lock.Unlock()
}

func (c *MapChannel) Close() {
	c.lock.Lock()
	for k, v := range c.channelDict {
		v.Close()
		delete(c.channelDict, k)
	}
	c.lock.Unlock()
}
