package transport

import (
	logger "EXSync/core/transport/logging"
	M "EXSync/core/transport/muxbuf"
	"context"
	"encoding/binary"
	"errors"
	"github.com/quic-go/quic-go"
	"io"
	"net"
	"sync/atomic"
)

// streamControlLength 流控制协议长度
// close-flag + 2 * streamID + control-flag
const (
	streamControlFlag   = 1
	streamControlLength = streamControlFlag + 2*streamIDLen + streamControlFlag
)

const (
	// defaultControlStream: create quic.Stream and delete quic.Stream
	// delete-flag, beforeStreamID, afterStreamID, reply-flag
	// format: {0/1 (type-flag) ,0-255, 0-255, 0-255, 0-255, 0-255, 0-255, 0-255, 1-255,
	// 0-255, 0-255, 0-255, 0-255, 0-255, 0-255, 0-255, 0-255, 0/1 (reply-flag) }
	defaultControlStream M.Mark = 1
)

var (
	// ErrCreateStream 会在初期发送流创建协议因Write函数遇到错误而返回
	ErrCreateStream = errors.New("create stream error")
	// ErrRecvStreamReply 遇到远程未答复流创建协议
	ErrRecvStreamReply = errors.New("recv stream reply error")
)

// getStream 向远程申请一个可用的流ID
// 该操作会请求创建流协议发送到远程默认控制流, 并等待一个是否可用答复
func (s *TCPConn) getStream(ctx context.Context, blockNum int) (quic.Stream, error) {
	// 获取一个基于自身流偏移的流ID
	beforeStreamID := s.createTCPStreamID()
	logger.Debugf("control-getStream-beforeStreamID: %v", beforeStreamID)
	// 创建流控制协议: 创建流 beforeStreamID
	scp, n := createStreamControlProto(streamControlProtocol{
		TypeFlag:       scpTypeFlagRequest,
		BeforeStreamID: beforeStreamID,
		AfterStreamID:  beforeStreamID,
		AckFlag:        1,
	})
	_, err := s.writeStream(scp[:n], defaultControlStream)
	if err != nil {
		return nil, ErrCreateStream
	}

	err = s.mc.createRecv(beforeStreamID)
	if err != nil {
		return nil, err
	}
	defer s.mc.del(beforeStreamID)

	var (
		mb            *M.MuxBuf
		afterStreamID M.Mark
		acceptFlag    byte
		ok            bool
	)

	for {
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		default:
			mb, ok = s.mc.getMuxBuf(beforeStreamID)
			if !ok {
				return nil, M.ErrMarkNotExist
			}
			n, err = s.mc.popTimeout(mb, &scp)
			if err != nil {
				return nil, ErrRecvStreamReply
			}

			streamControl := parseStreamControlProto(scp)

			beforeStreamID = streamControl.BeforeStreamID
			afterStreamID = streamControl.AfterStreamID
			acceptFlag = streamControl.AckFlag

			if acceptFlag == 1 {
				var stream quic.Stream
				stream, err = newTCPStream(afterStreamID, s)
				if err != nil {
					return nil, err
				}
				return stream, err
			}
			// 覆写请求
			afterStreamID = s.getOffsetStream(afterStreamID)
			binary.BigEndian.PutUint64(scp[n-1-2*streamIDLen:], afterStreamID)
			//M.CopyMarkToSlice(sc[n-1-2*streamIDLen:], afterStreamID)
			_, err = s.writeStream(scp[:n], beforeStreamID)
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

// closeStream 通知远程, 本地关闭的流
func (s *TCPConn) closeStream(closeMark M.Mark) error {
	sc, n := createStreamControlProto(streamControlProtocol{
		TypeFlag:       scpTypeFlagStream,
		BeforeStreamID: closeMark,
	})
	_, err := s.writeStream(sc[:n], defaultControlStream)
	if err != nil {
		return err
	}
	return nil
}

// createTCPStreamID 安全地获取一个未被占用的StreamID
func (s *TCPConn) createTCPStreamID() M.Mark {
	s.streamOffsetMutex.Lock()
	for s.mc.hasKey(s.streamOffset) {
		s.streamOffset += 1
		if s.streamOffset == ^uint64(0) { // 偏移调整归零
			s.streamOffset = 2
		}
	}
	s.streamOffsetMutex.Unlock()
	return s.streamOffset
}

// getOffsetStream 将原afterStreamID 根据本地与远程偏移，按顺序获取可用的流ID
func (s *TCPConn) getOffsetStream(afterStreamID M.Mark) M.Mark {
	remoteStreamOffset := afterStreamID
	if remoteStreamOffset > s.streamOffset {
		// 根据远程获取流偏移ID
		remoteStreamOffset += 1
		afterStreamID = remoteStreamOffset
		for s.mc.hasKey(afterStreamID) {
			if remoteStreamOffset == ^uint64(0) {
				remoteStreamOffset = 2
			}
			remoteStreamOffset += 1
			afterStreamID = remoteStreamOffset
		}
	} else {
		// 获取一个基于自身流偏移的流ID
		afterStreamID = s.createTCPStreamID()
	}
	return afterStreamID
}

// getControlStreamByte 阻塞并直到获取到一个 quic.Stream
func (s *TCPConn) getControlStreamByte(ctx context.Context, controlStream []byte) (stream quic.Stream, err error) {
	logger.Debugf("control-getControlStreamByte: %v", controlStream)
	logger.Debug("control-getControlStreamByte: enabled")
	mb, ok := s.mc.getMuxBuf(defaultControlStream)
	if !ok {
		return nil, M.ErrMarkNotExist
	}
	var n int
	select {
	case <-ctx.Done():
		logger.Debug("control-getControlStreamByte-done")
		return nil, context.Canceled
	default:
		logger.Debug("control-getControlStreamByte-default")
		n, err = s.mc.pop(mb, &controlStream)
		if err != nil {
			logger.Debugf("control-getControlStreamByte-err: %v", err)
			return nil, err
		}
	}
	logger.Debugf("control-getControlStreamByte: disabled %v", controlStream[:n])
	if n != streamControlLength {
		// 如果不是流创建协议则跳过
		return nil, ErrNonStreamControlProtocol
	}

	scp := parseStreamControlProto(controlStream[:n])
	stream, err = s.parseStream(scp)
	if err != nil {
		if err == net.ErrClosed {
			return s.getControlStreamByte(ctx, controlStream)
		}
		return nil, err
	}
	return stream, nil
}

// parseStream 处理一个远程有效的Stream到本地
// 创建默认流，默认流用于其它流的控制; streamID: 创建方根据未占用的StreamID指针选择, stat: 对上一个流的确认(0: false/1: true)
// 1. 远程发送自选择的创建流请求[id1, id1, stat(any)] // 创建请求会无视stat
// 2. 本地判断是否冲突, 是则响应[id2, stat(false)] // 冲突返回本地可接受的随机id, 并对上一个请求进行否定
// 3. 远程判断是否冲突, 是则沉默，否重新发送上述请求。 // 如不冲突则返回stat(true), 对上一个请求同意
func (s *TCPConn) parseStream(scp streamControlProtocol) (quic.Stream, error) {
	var err error
	switch scp.TypeFlag {
	case scpTypeFlagRequest:
		logger.Debugf("tcp_control-parseStream-scpRequest: %v", scp.BeforeStreamID)
		return s.createStreamProtocol(scp)
	case scpTypeFlagStream:
		logger.Debugf("tcp_control-parseStream-scpStreamClose: %v", scp.BeforeStreamID)
		if err = s.mc.del(scp.BeforeStreamID); err != nil {
			return nil, err
		}
		return nil, net.ErrClosed
	case scpTypeFlagConn:
		logger.Debugf("tcp_control-parseStream-scpConn_Code: %v", scp.AfterStreamID)
		return nil, s.closeLocal(quic.ApplicationErrorCode(scp.AfterStreamID), scp.ExtStr)
	default:
		return nil, errors.New("unknown streamCreateProtocol")
	}
}

func (s *TCPConn) createStreamProtocol(scp streamControlProtocol) (quic.Stream, error) {
	logger.Debugf("control-parseStream-before: %v", scp.BeforeStreamID)
	logger.Debugf("control-parseStream-after: %v", scp.AfterStreamID)
	var err error
	var stream quic.Stream
	if _, ok := s.taskMap[scp.BeforeStreamID]; ok {
		if s.mc.hasKey(scp.AfterStreamID) {
			if s.rejectNum > s.maxRejectNum {
				return nil, ErrStreamReject
			} else {
				atomic.AddUint64(&s.rejectNum, 1)
			}
			err = s.rejectStream(scp.BeforeStreamID, scp.AfterStreamID)
			return nil, err
		} else {
			stream, err = s.acceptStream(scp.BeforeStreamID, scp.AfterStreamID)
			if err != nil {
				return nil, err
			}
			return stream, nil
		}
	} else {
		if s.mc.hasKey(scp.AfterStreamID) {
			if s.rejectNum > s.maxRejectNum {
				err = ErrStreamReject
				return nil, err
			} else {
				atomic.AddUint64(&s.rejectNum, 1)
			}
			s.taskMap[scp.BeforeStreamID] = struct{}{}
			err = s.rejectStream(scp.BeforeStreamID, scp.AfterStreamID)
			return nil, err
		} else {
			// 同意时无视状态码
			logger.Debugf("control-parseStream-accept: %v, %v", scp.BeforeStreamID, scp.AfterStreamID)
			stream, err = s.acceptStream(scp.BeforeStreamID, scp.AfterStreamID)
			if err != nil {
				logger.Errorf("parseStream: %v", err)
				return nil, err
			}
			return stream, nil
		}
	}
}

// acceptStream 确认一个StreamID，并返回远程确认
func (s *TCPConn) acceptStream(beforeStreamID, afterStreamID M.Mark) (stream quic.Stream, err error) {
	defer func() {
		delete(s.taskMap, beforeStreamID)
	}()
	sc, n := createStreamControlProto(streamControlProtocol{
		TypeFlag:      scpTypeFlagRequest,
		AfterStreamID: afterStreamID,
		AckFlag:       1,
	})
	_, err = s.writeStream(sc[:n], beforeStreamID)
	if err != nil {
		return nil, err
	}

	// 创建接收通道
	err = s.mc.createRecv(afterStreamID)
	if err != nil {
		return nil, err
	}

	// 创建流
	return newTCPStream(afterStreamID, s)
}

// rejectStream 返回给流申请者一个本地能够接受的id
func (s *TCPConn) rejectStream(beforeStreamID, afterStreamID M.Mark) (err error) {
	afterStreamID = s.getOffsetStream(afterStreamID)
	sc, n := createStreamControlProto(streamControlProtocol{
		TypeFlag:      scpTypeFlagRequest,
		AfterStreamID: afterStreamID,
		AckFlag:       0,
	})
	_, err = s.writeStream(sc[:n], beforeStreamID)
	if err != nil {
		return err
	}
	return nil
}

// tcpStreamRecv TCP多路复用
// 收到数据包依次: 解粘包、解密(可选)、解压缩(可选)、放入MapChannel
func (s *TCPConn) tcpStreamRecv() {
	var (
		n              int
		ok             bool
		mark           M.Mark
		mb             *M.MuxBuf
		dataLen        uint16
		err            error
		compressDstBuf []byte

		dataLenBuf = make([]byte, 2)
		srcBuf     = make([]byte, socketLen)
		dstBuf     = make([]byte, socketLen)

		cipherLoss = getCipherLoss(s.cipher)
	)

	if s.compressor != nil {
		compressDstBuf = s.compressor.GetDstBuf()
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// 处理粘包
			n, err = s.conn.Read(dataLenBuf)
			if err != nil {
				if err == io.EOF {
					logger.Info("control-tcpStreamRecv-Read: ", err)
					return
				} else {
					logger.Warn(err)
				}
				return
			}

			dataLen = binary.BigEndian.Uint16(dataLenBuf)
			logger.Debugf("control-tcpStreamRecv-dataLen: %v", dataLen)
			if dataLen == 0 || dataLen > socketLen {
				continue
			}

			n, err = s.conn.Read(srcBuf[streamDataLen : dataLen+streamDataLen]) // 数据长度段不填充, 仅保留数据
			if err != nil {
				return
			}

			// 解密内容
			if s.cipher != nil {
				logger.Debugf("control-tcpStreamRecv-srcBuf: %v", srcBuf)
				logger.Debugf("control-tcpStreamRecv-srcBuf-de: %v, %v", dataLen, srcBuf[:dataLen])

				err = s.cipher.Decrypt(srcBuf[:dataLen], dstBuf)
				if err != nil {
					logger.Warnf("control-tcpStreamRecv-Decrypt: %v, %v", err, srcBuf[:dataLen])
					continue
				}
				logger.Debugf("control-tcpStreamRecv-dstBuf-tc: %v, %v", dataLen, dstBuf[:dataLen])
			} else {
				dstBuf = srcBuf
				logger.Debugf("control-tcpStreamRecv-dstBuf-noCipher: %v, %v", dataLen, dstBuf[:dataLen])
			}

			// 解压缩
			if s.compressor != nil {
				n, err = s.compressor.UnCompressData(dstBuf[streamDataLen+cipherLoss:dataLen],
					compressDstBuf)
				if err != nil {
					logger.Errorf("control-tcpStreamRecv-UnCompressData: %v, %v", err, dstBuf[streamDataLen+cipherLoss:dataLen])
					continue
				}
				logger.Debugf("control-tcpStreamRecv-dstBuf-tc: %v, %v", dataLen, dstBuf[:dataLen])
				mark = binary.BigEndian.Uint64(compressDstBuf[:n-1][:streamIDLen])
				dstBuf = compressDstBuf[streamIDLen:n]
			} else {
				logger.Debugf("control-tcpStreamRecv-mark: %v, %v", dataLen, dstBuf[:dataLen])
				mark = binary.BigEndian.Uint64(dstBuf[:dataLen][streamDataLen+cipherLoss : streamDataLen+cipherLoss+M.MarkLen])
				dstBuf = dstBuf[cipherLoss+streamDataLen+streamIDLen : dataLen]
			}

			// 发送到MapChannel
			logger.Infof("control-tcpStreamRecv-mark and dstBuf: %d, %v", mark, dstBuf)
			mb, ok = s.mc.getMuxBuf(mark)
			if !ok {
				logger.Errorf("tcpStreamRecv-getMuxBuf: %v: %v", M.ErrMarkNotExist, mark)
				continue
			}
			err = s.mc.push(mb, &dstBuf)
			if err != nil {
				logger.Errorf("tcpStreamRecv-pushCopy: %v", err)
				return
			}
		}
	}
}
