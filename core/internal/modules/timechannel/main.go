package timechannel

import (
	"EXSync/core/internal/config"
	"errors"
	"sync"
	"time"
)

type TimeChannel struct {
	channelDict map[string]chan []byte
	lock        sync.Mutex
}

// NewTimeChannel 创建一个数据接收队列
func NewTimeChannel() *TimeChannel {
	return &TimeChannel{
		make(map[string]chan []byte),
		sync.Mutex{},
	}
}

// HasKey 检查一个mark是否存在于time-channel中
func (t *TimeChannel) HasKey(mark string) (ok bool) {
	if _, ok := t.channelDict[mark]; ok {
		return true
	} else {
		return false
	}
}

// CreateRecv 创建一个数据流接收队列
func (t *TimeChannel) CreateRecv(mark string) (err error) {
	if _, ok := t.channelDict[mark]; !ok && len(mark) == 8 {
		t.channelDict[mark] = make(chan []byte, 65535)
		return nil
	}
	return errors.New("markExist")
}

// Set 在一个channel中写入值
func (t *TimeChannel) Set(mark string, value []byte) (ok bool) {
	if _, ok := t.channelDict[mark]; ok && len(mark) == 8 {
		channel := t.channelDict[mark]
		channel <- value
		return true
	} else {
		return false
	}
}

// Get 获取指定mark的首部，如果超时则返回timeout错误
func (t *TimeChannel) Get(mark string) (data []byte, err error) {
	if channel, ok := t.channelDict[mark]; ok {
		select {
		case value := <-channel:
			return value, nil
		case <-time.After(config.SocketTimeout * time.Second):
			return nil, errors.New("timeout")
		}
	}
	return nil, errors.New("markDoesNotExist")

}

// GetTimeout 获取指定mark的首部，如果超时则返回timeout错误
func (t *TimeChannel) GetTimeout(mark string, timeout int) (data []byte, err error) {
	if timeout == 0 {
		timeout = config.SocketTimeout
	}
	if channel, ok := t.channelDict[mark]; ok {
		select {
		case value := <-channel:
			return value, nil
		case <-time.After(time.Duration(timeout) * time.Second):
			return nil, errors.New("timeout")
		}
	}
	return nil, errors.New("markDoesNotExist")

}

// DelKey 释放指定mark的channel对象
func (t *TimeChannel) DelKey(mark string) {
	t.lock.Lock()
	if channel, ok := t.channelDict[mark]; ok {
		close(channel)
		delete(t.channelDict, mark)
	}
	t.lock.Unlock()
}

func (t *TimeChannel) Close() {
	t.lock.Lock()
	for k, v := range t.channelDict {
		close(v)
		delete(t.channelDict, k)
	}
	t.lock.Unlock()
}
