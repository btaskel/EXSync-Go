package commands

import (
	"EXSync/core/internal/exsync/server/commands/base"
	"EXSync/core/internal/exsync/server/commands/ext"
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	"EXSync/core/option"
	"github.com/sirupsen/logrus"
	"net"
)

type CommandProcess struct {
	AesGCM                    *encryption.Gcm
	TimeChannel               *timechannel.TimeChannel
	CommandSocket, DataSocket net.Conn
	IP                        string
	close                     bool
	VerifyManage              *map[string]option.VerifyManage
}

// NewCommandProcess 对已经被动建立连接的CommandSocket进行初始化
func NewCommandProcess(key string, dataSocket, commandSocket net.Conn, verifyManage *map[string]option.VerifyManage) {
	gcm, err := encryption.NewGCM(key)
	if err != nil {
		logrus.Errorf("NewCommandProcess: Error creating instruction processor using %s! %s", key, err)
		return
	}
	timeChannel := timechannel.NewTimeChannel()
	cp := CommandProcess{
		AesGCM:        gcm,
		TimeChannel:   timeChannel,
		DataSocket:    dataSocket,
		CommandSocket: commandSocket,
		VerifyManage:  verifyManage,
		close:         false,
	}
	cp.recvCommand()

	return
}

// recvCommand 以dict格式接收指令:
//
//	[8bytesMark]{
//	    "command": "data"/"comm", # 命令类型
//	    "type": "file",      # 操作类型
//	    "method": "get",     # 操作方法
//	    "data": {            # 参数数据集
//	        "a": 1
//	        ....
//	    }
//	}
func (c *CommandProcess) recvCommand() {
	commandSet := ext.CommandSet{Base: base.CommandSet{
		AesGCM:        c.AesGCM,
		Ip:            c.IP,
		TimeChannel:   c.TimeChannel,
		DataSocket:    c.DataSocket,
		CommandSocket: c.CommandSocket,
		VerifyManage:  c.VerifyManage,
	}}
	buf := make([]byte, 4096) // 数据接收切片
	for {
		if c.close {
			return
		}
		n, err := c.DataSocket.Read(buf)
		if err != nil {
			continue
		}

		if c.AesGCM == nil {
			c.recvNoEncrypt(n, buf)
		} else {
			c.recvAesGCM(n, buf)
		}
	}
}

// 接收使用Aes-GCM加密的数据
func (c *CommandProcess) recvAesGCM(n int, buf []byte) {
	if n <= 8 {
		// 错误数据
		return
	} else {
		// 未加密数据
		c.TimeChannel.Set(string(buf[:8]), buf[8:n])
		return
	}
}

// 接收未加密的数据
func (c *CommandProcess) recvNoEncrypt(n int, buf []byte) {
	if n <= 28 {
		// 不包含tag数据，无法验证其是否完整
		return
	} else {
		// 解密数据并判断数据是否有效并保存值timeChannel
		decryptData, err := c.AesGCM.AesGcmDecrypt(buf[:n])
		if err != nil {
			// 接收到无法解密的数据包
			return
		}
		data := decryptData[:8]
		mark := decryptData[8:]
		c.TimeChannel.Set(string(mark), data)
		return
	}
}
