package transport

import (
	logger "EXSync/core/log"
	"EXSync/core/transport/leakybuf"
	"encoding/binary"
	"errors"
)

var (
	taskMap = make(map[string]struct{}) // key: 初次StreamID
)

// getTCPStreamID 获取一个未被占用的StreamID
func (c *TCPConn) getTCPStreamID() ([]byte, error) {
	slice := make([]byte, streamIDSize)

	c.mutex.Lock()
	binary.BigEndian.PutUint32(slice, c.streamOffset)
	for c.mc.HasKey(string(slice)) {
		c.streamOffset += 1
		binary.BigEndian.PutUint32(slice, c.streamOffset)
	}
	c.mutex.Unlock()

	return slice, nil
}

func (c *TCPConn) getStream() {
	// todo:
}

// parseStream 处理一个远程有效的Stream到本地
// 创建默认流，默认流用于其它流的创建; id: 创建方随机选择, stat: 对上一个流的确认(0: false/1: true/2: any)
// 1. 远程发送自选择的创建流请求[id1, stat(any)] // 创建请求会无视stat
// 2. 本地判断是否冲突, 是则响应[id2, stat(false)] // 冲突返回本地可接受的随机id, 并对上一个请求进行否定
// 3. 远程判断是否冲突, 是则沉默，否重新发送上述请求。 // 如不冲突则返回stat(true), 对上一个请求同意
func (c *TCPConn) parseStream() (stream Stream, err error) {
	cs := make([]byte, c.socketDataLen)

	// 持续接收流申请
	var n int
	n, c.err = c.mc.Pull(string(defaultControlStream), &cs)
	if c.err != nil {
		if c.err == leakybuf.ErrTimeout {
			return
		}
		// 如果为除读取超时以外的错误，则退出默认流处理
		logger.Infof("Transport-TCP-initDefaultStream: %s: %s", c.conn.RemoteAddr(), c.err)
		return
	}

	if n != 9 {
		// 如果不是流创建协议则跳过
		return
	}

	beforeStreamID := cs[:4]
	afterStreamID := cs[4:8]
	streamIDStat := cs[8]

	if _, ok := taskMap[string(beforeStreamID)]; ok && streamIDStat == 1 {
		if c.rejectNum > c.maxRejectNum {
			c.err = errors.New("stream reject error")
			return
		} else {
			c.maxRejectNum += 1
		}
		err = c.rejectStream(beforeStreamID, afterStreamID)
		if err != nil {
			return
		}
	} else {
		taskMap[string(beforeStreamID)] = struct{}{}
	}

	if c.mc.HasKey(string(afterStreamID)) {
		if c.rejectNum > c.maxRejectNum {
			c.err = errors.New("stream reject error")
			return
		} else {
			c.maxRejectNum += 1
		}
		err = c.rejectStream(beforeStreamID, afterStreamID)
		if err != nil {
			return
		}
	} else {
		// 同意时无视状态码
		stream, err = c.acceptStream(beforeStreamID, afterStreamID)
		if err != nil {
			logger.Errorf("parseStream: %v", err)
			return
		}
		return stream, nil
	}
	return
}

// acceptStream 确认一个StreamID，并返回远程确认
func (c *TCPConn) acceptStream(beforeStreamID, afterStreamID []byte) (stream Stream, err error) {
	defer func() {
		delete(taskMap, string(beforeStreamID))
	}()

	idByte := make([]byte, streamLenSize+3*streamIDSize+1)

	// streamID
	idByte[2] = 0
	idByte[3] = 0
	idByte[4] = 0
	idByte[5] = 1

	// beforeStreamID 6-9
	for n, value := range beforeStreamID {
		idByte[n+6] = value
	}

	// afterStreamID 9-13
	for n, value := range afterStreamID {
		idByte[n+10] = value
	}

	// ackStream
	idByte[14] = 1

	_, err = c.Write(idByte, defaultControlStream)
	if err != nil {
		return nil, err
	}

	stream, err = newTCPStream(&c.tcpWithCipher, string(afterStreamID))
	if err != nil {
		return nil, err
	}

	return stream, nil
}

// rejectStream 返回给流申请者一个本地能够接受的id
func (c *TCPConn) rejectStream(beforeStreamID, afterStreamID []byte) (err error) {
	remoteStreamOffset := binary.BigEndian.Uint32(afterStreamID)
	if remoteStreamOffset > c.streamOffset {
		remoteStreamOffset += 1
		binary.BigEndian.PutUint32(afterStreamID, remoteStreamOffset)
		for c.mc.HasKey(string(afterStreamID)) {
			remoteStreamOffset += 1
			binary.BigEndian.PutUint32(afterStreamID, remoteStreamOffset)
		}
	} else {
		// 获取一个基于自身流偏移的流ID
		afterStreamID, err = c.getTCPStreamID()
		if err != nil {
			return err
		}
	}

	idByte := make([]byte, streamLenSize+3*streamIDSize+1)

	// streamID
	idByte[2] = 0
	idByte[3] = 0
	idByte[4] = 0
	idByte[5] = 1

	// beforeStreamID 6-9
	for n, value := range beforeStreamID {
		idByte[n+6] = value
	}

	// afterStreamID 9-13
	for n, value := range afterStreamID {
		idByte[n+10] = value
	}

	// ackStream
	idByte[14] = 0

	_, err = c.Write(idByte, defaultControlStream)
	if err != nil {
		return err
	}
	return nil
}

// tcpStreamRecv TCP多路复用
func (c *TCPConn) tcpStreamRecv() {
	defer func() {
		c.err = errors.New("tcpStreamRecv stopped")
	}()

	dataBuf := make([]byte, SocketSize)
	dataLenBuf := make([]byte, 2)
	var n int
	var dataLen uint16

	compressDstBuf := c.compressor.GetDstBuf()
	for {
		// 处理粘包
		_, c.err = c.conn.Read(dataLenBuf)
		if c.err != nil {
			logger.Debug(c.err)
			return
		}

		dataLen = binary.BigEndian.Uint16(dataLenBuf)
		n, c.err = c.conn.Read(dataBuf[:dataLen])
		if c.err != nil {
			return
		}

		// 解密内容
		if c.cipher != nil {
			c.err = c.cipher.Decrypt(dataBuf[:dataLen])
			if c.err != nil {
				continue
			}
		}

		// 解压缩
		if c.compressor != nil {
			n, c.err = c.compressor.UnCompressData(dataBuf[:dataLen], compressDstBuf)
			if c.err != nil {
				return
			}
			c.err = c.mc.Push(string(compressDstBuf[:n][:4]), compressDstBuf[4:n])
			if c.err != nil {
				logger.Errorf("tcpStreamRecv %s->%s: timeout!", c.conn.RemoteAddr(), c.conn.LocalAddr())
				return
			}
		} else {
			c.err = c.mc.Push(string(dataBuf[:dataLen][:4]), dataBuf[:dataLen][4:])
			if c.err != nil {
				logger.Errorf("tcpStreamRecv %s->%s: timeout!", c.conn.RemoteAddr(), c.conn.LocalAddr())
				return
			}
		}
	}
}
