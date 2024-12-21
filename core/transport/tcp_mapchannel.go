package transport

import (
	logger "EXSync/core/transport/logging"
	M "EXSync/core/transport/muxbuf"
	"fmt"
	"sync"
)

const (
	// mapChannelMaxChanLen 创建一个指定队列长度的网络接收缓冲区,
	// 最终内存占用由 mapChannelMaxChanLen * (socketLen - cipherLoss -
	// compressorLoss - streamDataLen - streamIDLen) 来决定
	mapChannelMaxChanLen = 32
)

type MapChannel struct {
	mapChan map[M.Mark]*M.MuxBuf
	lock    sync.Mutex
	bufLen  int
}

// newTimeChannel 创建一个数据接收队列, 每个队列默认最大使用1MB内存
func newTimeChannel(bufLen int) *MapChannel {
	return &MapChannel{
		mapChan: make(map[M.Mark]*M.MuxBuf, 8),
		bufLen:  bufLen,
	}
}

// getMuxBuf 获取一个MuxBuf
func (s *MapChannel) getMuxBuf(mark M.Mark) (*M.MuxBuf, bool) {
	muxBuf, ok := s.mapChan[mark]
	return muxBuf, ok
}

// hasKey 检查一个mark是否存在于MapChannel中
func (s *MapChannel) hasKey(mark M.Mark) (ok bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok = s.mapChan[mark]
	return ok
}

// createRecv 创建一个数据流接收队列
func (s *MapChannel) createRecv(mark M.Mark) (err error) {
	if !s.hasKey(mark) {
		s.mapChan[mark] = M.NewMuxBuf(mapChannelMaxChanLen, s.bufLen)
		return nil
	}
	return M.ErrMarkExist
}

func (s *MapChannel) pushCopy(muxBufPickFunc func() (*[]byte, error), putUsedFunc func(*[]byte), buf *[]byte) error {
	pushPointer, err := muxBufPickFunc()
	if err != nil {
		return err
	}
	fmt.Printf("pushCopy-pushPointer %p \n", pushPointer)
	*pushPointer = (*pushPointer)[:len(*buf)] // 调整自由切片大小为当前写入切片大小
	copy(*pushPointer, *buf)
	putUsedFunc(pushPointer)
	return nil
}

// pushTimeout 在一个channel中写入值, 如果关闭了muxBuf则返回EOF
func (s *MapChannel) pushTimeout(muxBuf *M.MuxBuf, buf *[]byte) error {
	return s.pushCopy(muxBuf.PickFreeTimeout, muxBuf.PutUsed, buf)
}

// Push 在一个muxBuf中写入缓冲
func (s *MapChannel) push(muxBuf *M.MuxBuf, buf *[]byte) error {
	return s.pushCopy(muxBuf.PickFree, muxBuf.PutUsed, buf)
}

func (s *MapChannel) popCopy(mb *M.MuxBuf, muxBufFunc func() (*[]byte, error), buf *[]byte) (int, error) {
	var n, remaining int
	var err error
	var popPointer *[]byte

	bufLen := len(*buf)

	var loopCounter int
	loopCounterEnd := mb.GetUsedBufLen()

	readBufferFully := func() (int, error) {
		// 判断当前切片是否超过缓冲区剩余空间
		if len(*popPointer) > remaining {
			// 部分复制数据到缓冲区
			n += copy((*buf)[n:], (*popPointer)[:remaining])
			// 存储未复制完的部分
			mb.SetSwapBuf((*popPointer)[remaining:])
			logger.Debugf("mapchannel-popCopy-*buf: %v", *buf)
			return n, nil
		}

		// 完全复制切片数据到缓冲区
		n += copy((*buf)[n:], *popPointer)
		return n, nil
	}

	for {
		logger.Debugf("mapchannel-popCopy-loopCounterEnd: %v", loopCounterEnd)

		// 获取数据切片
		popPointer, err = muxBufFunc()
		if err != nil {
			return 0, err
		}

		logger.Debugf("mapchannel-popCopy-popPointer: %v", popPointer)

		// 计算剩余缓冲区大小
		remaining = bufLen - n

		if loopCounterEnd == 0 {
			return readBufferFully()
		}

		n, err = readBufferFully()
		if err != nil {
			return 0, err
		}

		// 如果缓冲区刚好填满，返回结果
		if n == bufLen {
			logger.Debugf("mapchannel-popCopy-*buf: %v", *buf)
			return n, nil
		}

		loopCounter++
		if loopCounter == loopCounterEnd {
			return n, nil
		}
	}
}

// popTimeout 受 SetReadDeadline 控制, 获取指定mark的数据
// 如果超时则返回timeout错误, 如果关闭了muxBuf则返回io.EOF
func (s *MapChannel) popTimeout(mb *M.MuxBuf, buf *[]byte) (int, error) {
	return s.popCopy(mb, mb.PickUsedTimeout, buf)
}

// Pop 获取指定mark的数据，如果超时则返回timeout错误
// 如果关闭了muxBuf则返回io.EOF
func (s *MapChannel) pop(mb *M.MuxBuf, buf *[]byte) (int, error) {
	return s.popCopy(mb, mb.PickUsed, buf)
}

// del 释放指定mark的channel对象, 在对 MuxBuf 使用 popCopy
// 函数时将会取出剩下的byte数据, 读取完毕后返回 io.EOF
func (s *MapChannel) del(mark M.Mark) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if mb, ok := s.mapChan[mark]; ok {
		mb.Close()
		delete(s.mapChan, mark)
		return nil
	}
	return M.ErrMarkNotExist
}

// close 关闭 MapChannel 并释放所有资源。
// 它会锁定 MapChannel，关闭所有 MuxBuf 并从 mapChan 中删除它们。
func (s *MapChannel) close() {
	logger.Warnf("mapChannel closed")
	s.lock.Lock()
	defer s.lock.Unlock()
	for mark, mb := range s.mapChan {
		mb.Close()
		delete(s.mapChan, mark)
	}
}
