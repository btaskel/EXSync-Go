package commands

import (
	"EXSync/core/internal/exsync/server/commands/ext"
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	"EXSync/core/option"
	"EXSync/core/option/server/comm"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"net"
)

type CommandProcess struct {
	AesGCM        *encryption.Gcm
	TimeChannel   *timechannel.TimeChannel
	CommandSocket net.Conn
	DataSocket    net.Conn
	IP            string
	VerifyManage  *map[string]option.VerifyManage

	close bool
}

// NewCommandProcess 对已经被动建立连接的CommandSocket进行初始化
func NewCommandProcess(host string, dataSocket, commandSocket net.Conn, verifyManage *map[string]option.VerifyManage) {
	var gcm *encryption.Gcm
	hostVerifyInfo, ok := (*verifyManage)[host]
	if ok && len(hostVerifyInfo.AesKey) != 0 {
		var err error
		gcm, err = encryption.NewGCM(hostVerifyInfo.AesKey)
		if err != nil {
			logrus.Errorf("NewCommandProcess: Error creating instruction processor using %s! %s", hostVerifyInfo.AesKey, err)
			return
		}
	} else {
		gcm = nil
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

// recvCommand  创建指令接收队列, 以map格式接收指令:
func (c *CommandProcess) recvCommand() {
	commandSet, err := ext.NewCommandSet(c.TimeChannel, c.DataSocket, c.CommandSocket, c.AesGCM, c.VerifyManage, 28)
	if err != nil {
		logrus.Errorf("recvCommand: Failed to initialize commandSet! %s", err)
		return
	}
	go c.recvData()
	var command comm.Command
	buf := make([]byte, 4096)
	for {
		if c.close {
			return
		}
		n, err := c.CommandSocket.Read(buf)
		if err != nil {
			return
		}
		result, err := c.decryptData(n, buf)
		if err != nil {
			continue
		}
		err = json.Unmarshal(result[8:], &command)
		if err != nil {
			continue
		}
		go commandSet.MatchCommand(command)
	}
}

// recvData 创建数据接收队列
func (c *CommandProcess) recvData() {
	buf := make([]byte, 4096) // 数据接收切片
	for {
		if c.close {
			return
		}
		n, err := c.DataSocket.Read(buf)
		if err != nil {
			continue
		}

		result, err := c.decryptData(n, buf)
		c.TimeChannel.Set(string(result[:8]), result[8:])
	}
}

// decryptData 判断当前会话数据是否进行了加密，并尝试解密
func (c *CommandProcess) decryptData(n int, buf []byte) (data []byte, err error) {
	if c.AesGCM == nil {
		if n <= 8 {
			// 错误数据
			return nil, errors.New("数据长度小于8")
		} else {
			// 未加密数据
			return buf, nil
		}
	} else {
		if n <= 28 {
			// 不包含tag数据，无法验证其是否完整
			return nil, errors.New("数据长度小于28")
		} else {
			// 解密数据并判断数据是否有效并保存值timeChannel
			decryptData, err := c.AesGCM.AesGcmDecrypt(buf[:n])
			if err != nil {
				// 接收到无法解密的数据包
				return nil, err
			}
			return decryptData, nil
		}
	}
}
