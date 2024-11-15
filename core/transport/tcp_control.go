package transport

import (
	logger "EXSync/core/log"
	M "EXSync/core/transport/muxbuf"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// streamControlLength 流控制协议长度
// close-flag + 2 * streamID + control-flag
const streamControlLength = 1 + 2*streamIDLen + 1

var (
	// ErrCreateStream 会在初期发送流创建协议因Write函数遇到错误而返回
	ErrCreateStream = errors.New("create stream error")
	// ErrCreateStreamReply 遇到远程未答复流创建协议
	ErrCreateStreamReply = errors.New("create stream reply error")
)

// getStream 向远程申请一个可用的流ID
// 该操作会请求创建流协议发送到远程默认控制流, 并等待一个是否可用答复
func (c *TCPConn) getStream(ctx context.Context, blockNum int) (Stream, error) {
	// 获取一个基于自身流偏移的流ID
	beforeStreamID := c.createTCPStreamID()
	fmt.Println("getStream-beforeStreamID: ", beforeStreamID)
	// 创建流控制协议: 创建流 beforeStreamID
	sc, n := c.createStreamControlProto(streamControlOption{
		CloseFlag:      0,
		BeforeStreamID: beforeStreamID,
		AfterStreamID:  beforeStreamID,
		AckFlag:        1,
	})
	_, err := c.writeStream(sc[:n], defaultControlStream)
	if err != nil {
		return nil, ErrCreateStream
	}

	err = c.mc.CreateRecv(beforeStreamID)
	if err != nil {
		return nil, err
	}
	defer c.mc.Del(beforeStreamID)

	var (
		mb            *M.MuxBuf
		afterStreamID M.Mark
		acceptFlag    byte
		ok            bool
	)
	counter := 0
	for {
		fmt.Println("counter:", counter)
		counter += 1
		select {
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		default:
			fmt.Println("pop start", beforeStreamID)
			mb, ok = c.mc.GetMuxBuf(beforeStreamID)
			if !ok {
				return nil, M.ErrMarkNotExist
			}
			n, err = c.mc.PopTimeout(mb, &sc)
			if err != nil {
				return nil, ErrCreateStreamReply
			}
			fmt.Println("pop end")

			fmt.Println("parseStreamControl", sc)
			streamControl := c.parseStreamControl(sc)

			beforeStreamID = streamControl.BeforeStreamID
			afterStreamID = streamControl.AfterStreamID
			acceptFlag = streamControl.AckFlag

			if acceptFlag == 1 {
				var stream Stream
				stream, err = newTCPStream(c.tcpWithCipher, afterStreamID, c)
				if err != nil {
					return nil, err
				}
				return stream, err
			} else {
				// 覆写请求
				afterStreamID = c.getOffsetStream(afterStreamID)
				M.CopyMarkToSlice(sc[n-1-2*streamIDLen:], afterStreamID)
				_, err = c.writeStream(sc[:n], beforeStreamID)
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

type streamControlOption struct {
	CloseFlag                     byte
	BeforeStreamID, AfterStreamID M.Mark
	AckFlag                       byte
}

// createStreamControlProto 创建一个流控制协议
func (c *TCPConn) createStreamControlProto(streamControlOption streamControlOption) ([]byte, int) {
	sc, n := c.getSocketSlice()

	// closeFlag 0
	sc[n+0] = streamControlOption.CloseFlag

	// beforeStreamID 1-4
	for i, value := range streamControlOption.BeforeStreamID {
		sc[n+1+i] = value
	}

	// afterStreamID 5-
	for i, value := range streamControlOption.AfterStreamID {
		sc[n+5+i] = value
	}

	// ackFlag 9
	sc[n+9] = streamControlOption.AckFlag
	fmt.Println("sc:", sc)
	return sc, n + 10
}

func (c *TCPConn) parseStreamControl(sc []byte) streamControlOption {
	fmt.Println("parseStreamControl", sc)
	// [0 0 0 0 0 0 0 0 2 1]
	return streamControlOption{
		CloseFlag:      sc[0],
		BeforeStreamID: M.SliceToMark(sc[1:5]),
		AfterStreamID:  M.SliceToMark(sc[5:9]),
		AckFlag:        sc[9],
	}
}

// closeStream 关闭指定Mark的流
func (c *TCPConn) closeStream(closeMark M.Mark) error {
	sc, n := c.createStreamControlProto(streamControlOption{
		CloseFlag:      1,
		BeforeStreamID: closeMark,
	})
	_, err := c.writeStream(sc[:n], defaultControlStream)
	if err != nil {
		return err
	}
	return nil
}

// createTCPStreamID 安全地获取一个未被占用的StreamID
func (c *TCPConn) createTCPStreamID() M.Mark {
	c.streamOffsetMutex.Lock()
	arr := M.PutUint32(c.streamOffset)
	for c.mc.HasKey(arr) {
		c.streamOffset += 1
		if c.streamOffset == ^uint32(0) { // 偏移调整归零
			c.streamOffset = 2
		}
		arr = M.PutUint32(c.streamOffset)
	}
	c.streamOffsetMutex.Unlock()
	return arr
}

// getOffsetStream 将原afterStreamID 根据本地与远程偏移，按顺序获取可用的流ID
func (c *TCPConn) getOffsetStream(afterStreamID M.Mark) M.Mark {
	remoteStreamOffset := M.MarkToUint32(afterStreamID)
	if remoteStreamOffset > c.streamOffset {
		// 根据远程获取流偏移ID
		remoteStreamOffset += 1
		afterStreamID = M.PutUint32(remoteStreamOffset)
		for c.mc.HasKey(afterStreamID) {
			if remoteStreamOffset == ^uint32(0) {
				remoteStreamOffset = 2
			}
			remoteStreamOffset += 1
			afterStreamID = M.PutUint32(remoteStreamOffset)
		}
	} else {
		// 获取一个基于自身流偏移的流ID
		streamID := c.createTCPStreamID()
		for i := 0; i < 4; i++ {
			afterStreamID[i] = streamID[i]
		}
	}
	return afterStreamID
}

// getControlStreamByte 阻塞并捕获控制流
// 从MapChannel 中弹出一个流创建信息，并转交下层处理
func (c *TCPConn) getControlStreamByte(controlStream []byte) (stream Stream, err error) {
	// 持续接收流申请
	fmt.Println("getControlStreamByte: ", controlStream)
	var n int
	fmt.Println("getControlStreamByte: enabled")
	mb, ok := c.mc.GetMuxBuf(defaultControlStream)
	if !ok {
		return nil, M.ErrMarkNotExist
	}
	n, c.err = c.mc.Pop(mb, &controlStream)
	if c.err != nil {
		return nil, c.err
	}
	fmt.Println("getControlStreamByte: disabled", controlStream[:n])
	if n != streamControlLength {
		// 如果不是流创建协议则跳过
		return nil, ErrNonStreamControlProtocol
	}

	stream, err = c.parseStream(controlStream[:n])
	if err != nil {
		return nil, err
	}
	return stream, nil
}

// parseStream 处理一个远程有效的Stream到本地
// 创建默认流，默认流用于其它流的创建; streamID: 创建方随机选择, stat: 对上一个流的确认(0: false/1: true/2: any)
// 1. 远程发送自选择的创建流请求[id1, stat(any)] // 创建请求会无视stat
// 2. 本地判断是否冲突, 是则响应[id2, stat(false)] // 冲突返回本地可接受的随机id, 并对上一个请求进行否定
// 3. 远程判断是否冲突, 是则沉默，否重新发送上述请求。 // 如不冲突则返回stat(true), 对上一个请求同意
func (c *TCPConn) parseStream(cs []byte) (Stream, error) {
	closeFlag := cs[0]
	beforeStreamID := M.SliceToMark(cs[1:])
	afterStreamID := M.SliceToMark(cs[5:])

	if closeFlag == 1 {
		c.err = c.mc.Del(beforeStreamID)
		if c.err != nil {
			return nil, c.err
		}
		return nil, ErrStreamClosed
	}

	fmt.Println("before: ", beforeStreamID)
	fmt.Println("after: ", afterStreamID)
	if _, ok := c.taskMap[beforeStreamID]; ok {
		if c.mc.HasKey(afterStreamID) {
			c.rejectNumMutex.Lock()
			if c.rejectNum > c.maxRejectNum {
				c.err = ErrStreamReject
				return nil, c.err
			} else {
				c.maxRejectNum += 1
			}
			c.rejectNumMutex.Unlock()
			c.err = c.rejectStream(beforeStreamID, afterStreamID)
			return nil, c.err
		} else {
			var stream Stream
			stream, c.err = c.acceptStream(beforeStreamID, afterStreamID)
			if c.err != nil {
				return nil, c.err
			}
			return stream, nil
		}
	} else {
		if c.mc.HasKey(afterStreamID) {
			if c.rejectNum > c.maxRejectNum {
				c.err = ErrStreamReject
				return nil, c.err
			} else {
				c.maxRejectNum += 1
			}
			c.taskMap[beforeStreamID] = struct{}{}
			c.err = c.rejectStream(beforeStreamID, afterStreamID)
			return nil, c.err
		} else {
			// 同意时无视状态码
			var stream Stream
			fmt.Println("accept: ", beforeStreamID, afterStreamID)
			stream, c.err = c.acceptStream(beforeStreamID, afterStreamID)
			if c.err != nil {
				logger.Errorf("parseStream: %v", c.err)
				return nil, c.err
			}
			return stream, nil
		}
	}
}

// acceptStream 确认一个StreamID，并返回远程确认
func (c *TCPConn) acceptStream(beforeStreamID, afterStreamID M.Mark) (stream Stream, err error) {
	defer func() {
		delete(c.taskMap, beforeStreamID)
	}()
	sc, n := c.createStreamControlProto(streamControlOption{
		CloseFlag:     0,
		AfterStreamID: afterStreamID,
		AckFlag:       1,
	})
	_, err = c.writeStream(sc[:n], beforeStreamID)
	if err != nil {
		return nil, err
	}

	// 创建接收通道
	err = c.mc.CreateRecv(afterStreamID)
	if err != nil {
		return nil, err
	}

	// 创建流
	stream, err = newTCPStream(c.tcpWithCipher, afterStreamID, c)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

// rejectStream 返回给流申请者一个本地能够接受的id
func (c *TCPConn) rejectStream(beforeStreamID, afterStreamID M.Mark) (err error) {
	afterStreamID = c.getOffsetStream(afterStreamID)
	sc, n := c.createStreamControlProto(streamControlOption{
		CloseFlag:     0,
		AfterStreamID: afterStreamID,
		AckFlag:       0,
	})
	_, err = c.writeStream(sc[:n], beforeStreamID)
	if err != nil {
		return err
	}
	return nil
}

// tcpStreamRecv TCP多路复用
// 收到数据包依次: 解粘包、解密(可选)、解压缩(可选)、放入MapChannel
func (c *TCPConn) tcpStreamRecv() {
	defer func() {
		c.err = errors.New("tcpStreamRecv stopped")
	}()

	var (
		n              int
		ok             bool
		mark           M.Mark
		mb             *M.MuxBuf
		dataLen        uint16
		compressDstBuf []byte

		dataLenBuf = make([]byte, 2)
		srcBuf     = make([]byte, socketLen)
		dstBuf     = make([]byte, socketLen)

		cipherLoss = getCipherLoss(c.cipher)
	)

	if c.compressor != nil {
		compressDstBuf = c.compressor.GetDstBuf()
	}

	for {
		// 处理粘包
		n, c.err = c.conn.Read(dataLenBuf)
		if c.err != nil {
			if c.err == io.EOF {
				logger.Info("tcpStreamRecv: ", c.err)
				return
			} else {
				logger.Error(c.err)
			}
			return
		}

		dataLen = binary.BigEndian.Uint16(dataLenBuf)

		if dataLen == 0 || dataLen > socketLen {
			continue
		}

		n, c.err = c.conn.Read(srcBuf[streamDataLen : dataLen+streamDataLen]) // 数据长度段不填充, 仅保留数据
		if c.err != nil {
			return
		}

		// 解密内容
		if c.cipher != nil {
			fmt.Println("srcBuf", srcBuf)
			fmt.Println("srcBuf-de", dataLen, srcBuf[:dataLen])
			c.err = c.cipher.Decrypt(srcBuf[:dataLen], dstBuf)
			if c.err != nil {
				fmt.Println("tcpStreamRecv-Decrypt", c.err, srcBuf[:dataLen])
				continue
			}
			fmt.Println("dstBuf-tc", dataLen, dstBuf[:dataLen])
		} else {

		}

		// 解压缩
		if c.compressor != nil {
			n, c.err = c.compressor.UnCompressData(dstBuf[streamDataLen+cipherLoss:dataLen],
				compressDstBuf)
			if c.err != nil {
				fmt.Println("tcpStreamRecv-UnCompressData: ", c.err, dstBuf[streamDataLen+cipherLoss:dataLen])
				logger.Errorf("tcpStreamRecv-UnCompressData: ", c.err)
				return
			}
			fmt.Println("compressDstBuf", compressDstBuf[:n-1][:streamIDLen])
			mark = M.SliceToMark(compressDstBuf[:n-1][:streamIDLen])
			dstBuf = compressDstBuf[streamIDLen:n]
		} else {
			fmt.Println("de-mark", dstBuf[:dataLen])
			mark = M.SliceToMark(dstBuf[:dataLen][streamDataLen+cipherLoss : streamDataLen+cipherLoss+4])
			dstBuf = dstBuf[cipherLoss+streamDataLen+streamIDLen : dataLen]
		}

		// 发送到MapChannel
		fmt.Println("mark and dstBuf", mark, dstBuf)

		mb, ok = c.mc.GetMuxBuf(mark)
		if !ok {
			logger.Errorf("tcpStreamRecv-GetMuxBuf: ", M.ErrMarkNotExist)
			continue
		}
		c.err = c.mc.Push(mb, &dstBuf)
		if c.err != nil {
			logger.Errorf("tcpStreamRecv-push: ", c.err)
			return
		}
		fmt.Println("push-mark", mark, dstBuf)
	}
}
