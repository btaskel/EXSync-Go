package transport

import (
	logger "EXSync/core/transport/logging"
	M "EXSync/core/transport/muxbuf"
	"encoding/binary"
)

type scpByte = byte

const (
	// scpTypeFlagRequest 流申请协议
	scpTypeFlagRequest scpByte = iota
	// scpTypeFlagStream 流关闭协议
	scpTypeFlagStream
	// scpTypeFlagConn 套接字关闭请求
	scpTypeFlagConn
)

// streamControlProtocol 它通过流: defaultControlStream
// 作用于对 Stream 与 quic.Connection 的控制
type streamControlProtocol struct {
	// TypeFlag = 0, 流申请协议, BeforeStreamID 用于开启的目标流ID, AfterStreamID 本次请求可以接受的流ID
	// TypeFlag = 1, 流关闭协议, BeforeStreamID 是关闭的目标流ID, 其余忽视
	// TypeFlag = 2, 套接字关闭请求, BeforeStreamID, AfterStreamID 则为结束代码, 其余忽视
	//todo: TypeFlag = 3, 数据报流协议
	TypeFlag scpByte
	// BeforeStreamID 通常是作为作用的目标流
	BeforeStreamID M.Mark
	// AfterStreamID 通常是作为作用的目标流code
	AfterStreamID M.Mark
	// AckFlag 对上一个流行为的同意
	AckFlag byte
	// ExtStr 可拓展的字符串描述
	ExtStr string
}

// createStreamControlProto 创建一个 []byte 类型的流控制协议
// scp 是一个 streamControlProtocol 结构体，包含流控制协议信息
// 返回一个字节切片，表示流控制协议的数据，以及数据的长度
func createStreamControlProto(scp streamControlProtocol) ([]byte, int) {
	extStrSlice := []byte(scp.ExtStr)
	logger.Warnf("createStreamControlProto: %v", extStrSlice)
	sc := make([]byte, streamControlLength+len(extStrSlice))

	sc[0] = scp.TypeFlag

	binary.BigEndian.PutUint64(sc[1:], scp.BeforeStreamID)

	binary.BigEndian.PutUint64(sc[1+M.MarkLen:], scp.AfterStreamID)

	sc[1+M.MarkLen*2] = scp.AckFlag

	copy(sc[1+M.MarkLen*2+1:], extStrSlice)

	logger.Debug("control-createStreamControlProto-sc: ", sc)
	return sc, streamControlLength + len(extStrSlice)
}

// parseStreamControlProto 解析一个流控制协议
// sc 是一个字节切片，表示流控制协议的数据
// 返回一个 streamControlProtocol 结构体，包含解析后的协议信息
func parseStreamControlProto(sc []byte) streamControlProtocol {
	logger.Debug("control-parseStreamControlProto-sc: ", sc)
	scp := streamControlProtocol{
		// TypeFlag 表示协议类型
		TypeFlag: sc[0],
		// BeforeStreamID 表示作用的目标流ID
		BeforeStreamID: binary.BigEndian.Uint64(sc[1 : 1+M.MarkLen]),
		// AfterStreamID 表示作用的目标流ID
		AfterStreamID: binary.BigEndian.Uint64(sc[1+M.MarkLen : 1+M.MarkLen*2]),
		// AckFlag 表示对上一个流行为的同意
		AckFlag: sc[1+M.MarkLen*2],
	}
	// 如果字节切片长度不等于 streamControlLength，则解析 ExtStr
	if len(sc) != streamControlLength {
		scp.ExtStr = string(sc[1+M.MarkLen*2+1:])
	}
	return scp
}
