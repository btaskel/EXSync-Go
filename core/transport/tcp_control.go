package transport

import (
	logger "EXSync/core/log"
	"EXSync/core/transport/leakybuf"
	"context"
	"encoding/binary"
	"errors"
)

const uint32Max = ^uint32(0)

// getStream 向远程申请一个可用的流ID
// 该操作讲请求创建流请求发送到默认控制流
func (c *TCPConn) getStream(ctx context.Context, blockNum int) (Stream, error) {
	idByte := make([]byte, streamLenSize+3*streamIDSize+1)

	// 获取一个基于自身流偏移的流ID
	beforeStreamID, err := c.getTCPStreamID()
	if err != nil {
		return nil, err
	}

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
	for n, value := range beforeStreamID {
		idByte[n+10] = value
	}

	// ackStream
	idByte[14] = 1

	_, err = c.Write(idByte, defaultControlStream)
	if err != nil {
		return nil, err
	}

	err = c.mc.CreateRecv(string(beforeStreamID))
	defer c.mc.Del(string(beforeStreamID))
	if err != nil {
		return nil, err
	}

	var n int
	buf := make([]byte, c.socketDataLen)

	for {
		select {
		case <-ctx.Done():
			return nil, nil
		default:
			n, err = c.mc.PopTimeout(string(beforeStreamID), &buf)
			if err != nil {
				return nil, err
			}

			beforeStreamID = buf[:n][:4]
			afterStreamID := buf[:n][4:8]
			acceptFlag := buf[:n][8]

			if acceptFlag == 1 {
				var stream Stream
				stream, err = newTCPStream(c.tcpWithCipher, string(afterStreamID))
				if err != nil {
					return stream, err
				}
			} else {
				// 覆写请求

				err = c.getOffsetStream(afterStreamID)
				if err != nil {
					return nil, err
				}

				_, err = c.Write(buf, beforeStreamID)
				if err != nil {
					return nil, err
				}

				if blockNum == -1 {
					continue
				} else {
					blockNum++
					if blockNum >= 5 {
						break
					}
				}
			}
		}
	}
}

// getTCPStreamID 安全地获取一个未被占用的StreamID
func (c *TCPConn) getTCPStreamID() ([]byte, error) {
	slice := make([]byte, streamIDSize)

	c.streamOffsetMutex.Lock()
	binary.BigEndian.PutUint32(slice, c.streamOffset)
	for c.mc.HasKey(string(slice)) {
		c.streamOffset += 1
		if c.streamOffset == uint32Max { // 偏移调整归零
			c.streamOffset = 0
		}
		binary.BigEndian.PutUint32(slice, c.streamOffset)
	}
	c.streamOffsetMutex.Unlock()

	return slice, nil
}

// getOffsetStream 将原afterStreamID 根据本地与远程偏移，按顺序获取可用的流ID
func (c *TCPConn) getOffsetStream(afterStreamID []byte) error {
	remoteStreamOffset := binary.BigEndian.Uint32(afterStreamID)
	if remoteStreamOffset > c.streamOffset {
		// 根据远程获取流偏移ID
		remoteStreamOffset += 1
		binary.BigEndian.PutUint32(afterStreamID, remoteStreamOffset)
		for c.mc.HasKey(string(afterStreamID)) {
			if remoteStreamOffset == uint32Max {
				remoteStreamOffset = 0
			}
			remoteStreamOffset += 1
			binary.BigEndian.PutUint32(afterStreamID, remoteStreamOffset)
		}
	} else {
		// 获取一个基于自身流偏移的流ID
		var err error
		afterStreamID, err = c.getTCPStreamID()
		if err != nil {
			return err
		}
	}
	return nil
}

// getControlStreamByte 阻塞并捕获控制流
// 从MapChannel 中弹出一个流创建信息，并转交下层处理
func (c *TCPConn) getControlStreamByte(controlStream []byte) (stream Stream, err error) {
	// 持续接收流申请
	var n int
	n, c.err = c.mc.Pop(string(defaultControlStream), &controlStream)
	if c.err != nil {
		if c.err == leakybuf.ErrTimeout {
			return nil, err
		}
		// 如果为除读取超时以外的错误，则退出默认流处理 nn
		logger.Infof("Transport-TCP-initDefaultStream: %s: %s", c.conn.RemoteAddr(), c.err)
		return nil, err
	}
	stream, err = c.parseStream(n, controlStream)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

// parseStream 处理一个远程有效的Stream到本地
// 创建默认流，默认流用于其它流的创建; id: 创建方随机选择, stat: 对上一个流的确认(0: false/1: true/2: any)
// 1. 远程发送自选择的创建流请求[id1, stat(any)] // 创建请求会无视stat
// 2. 本地判断是否冲突, 是则响应[id2, stat(false)] // 冲突返回本地可接受的随机id, 并对上一个请求进行否定
// 3. 远程判断是否冲突, 是则沉默，否重新发送上述请求。 // 如不冲突则返回stat(true), 对上一个请求同意
func (c *TCPConn) parseStream(n int, cs []byte) (stream Stream, err error) {
	if n != 2*streamIDSize+1 {
		// 如果不是流创建协议则跳过
		return
	}

	beforeStreamID := cs[:4]
	afterStreamID := cs[4:8]

	if _, ok := c.taskMap[string(beforeStreamID)]; ok {
		if c.mc.HasKey(string(afterStreamID)) {
			c.rejectNumMutex.Lock()
			if c.rejectNum > c.maxRejectNum {
				c.err = errors.New("stream reject error")
				return
			} else {
				c.maxRejectNum += 1
			}
			c.rejectNumMutex.Unlock()
			err = c.rejectStream(beforeStreamID, afterStreamID)
			if err != nil {
				return
			}
			return
		} else {
			var acceptStream Stream
			acceptStream, err = c.acceptStream(beforeStreamID, afterStreamID)
			if err != nil {
				return nil, err
			}
			return acceptStream, nil
		}
	} else {
		if c.mc.HasKey(string(afterStreamID)) {
			if c.rejectNum > c.maxRejectNum {
				c.err = errors.New("stream reject error")
				return
			} else {
				c.maxRejectNum += 1
			}
			c.taskMap[string(beforeStreamID)] = nil
			err = c.rejectStream(beforeStreamID, afterStreamID)
			if err != nil {
				return
			}
			return
		} else {
			// 同意时无视状态码
			stream, err = c.acceptStream(beforeStreamID, afterStreamID)
			if err != nil {
				logger.Errorf("parseStream: %v", err)
				return
			}
			return stream, nil
		}
	}
}

// acceptStream 确认一个StreamID，并返回远程确认
func (c *TCPConn) acceptStream(beforeStreamID, afterStreamID []byte) (stream Stream, err error) {
	defer func() {
		delete(c.taskMap, string(beforeStreamID))
	}()

	idByte := make([]byte, streamLenSize+2*streamIDSize+1)

	// afterStreamID 6-9
	for n, value := range afterStreamID {
		idByte[streamLenSize+streamIDSize+n] = value
	}

	// ackStream
	idByte[14] = 1

	_, err = c.Write(idByte, beforeStreamID)
	if err != nil {
		return nil, err
	}

	stream, err = newTCPStream(c.tcpWithCipher, string(afterStreamID))
	if err != nil {
		return nil, err
	}

	return stream, nil
}

// rejectStream 返回给流申请者一个本地能够接受的id
func (c *TCPConn) rejectStream(beforeStreamID, afterStreamID []byte) (err error) {
	err = c.getOffsetStream(afterStreamID)
	if err != nil {
		return err
	}

	idByte := make([]byte, streamLenSize+2*streamIDSize+1)

	// afterStreamID 6-9
	for n, value := range afterStreamID {
		idByte[streamLenSize+streamIDSize+n] = value
	}

	// ackStream
	idByte[10] = 0

	_, err = c.Write(idByte, beforeStreamID)
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
