package commands

import (
	"EXSync/core/internal/exsync/server"
	"EXSync/core/internal/exsync/server/commands/base"
	"EXSync/core/internal/exsync/server/commands/ext"
	loger "EXSync/core/log"
	"EXSync/core/option/exsync/comm"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"sync"
)

// recvCommand  创建指令接收队列, 以map格式接收指令:
func (c *CommandProcess) recvCommand(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	go func() {
		select {
		case <-c.ctxProcess.Done():
			loger.Log.Debugf("recvCommand cancel: %s", c.ctxProcess.Err())
			wg.Done()
			return
		}
	}()

	// 创建指令处理对象
	serverCommBase := base.Base{
		Ip:             c.IP,
		TimeChannel:    c.TimeChannel,
		DataSocket:     c.DataSocket,
		CommandSocket:  c.CommandSocket,
		AesGCM:         c.AesGCM,
		VerifyManage:   server.VerifyManage,
		EncryptionLoss: 28,
		CtxProcess:     c.ctxProcess,
		TaskManage:     c.taskManage,
	}
	commandSet, err := ext.NewCommandSet(serverCommBase)
	if err != nil {
		loger.Log.Errorf("recvCommand: Failed to initialize commandSet! %s", err)
		return
	}
	var command comm.Command
	for {
		lengthBuf := make([]byte, 2)
		_, err := io.ReadFull(c.CommandSocket, lengthBuf)
		if err != nil {
			if err == io.EOF {
				loger.Log.Debugf("Passive-recvCommand: recvCommand connection disconnected from host %s.", c.IP)
				return
			} else {
				continue
			}
		}

		buf := make([]byte, binary.BigEndian.Uint16(lengthBuf))
		n, err := c.CommandSocket.Read(buf)
		if err != nil {
			if err == io.EOF {
				loger.Log.Debugf("Passive-recvCommand: connection disconnected from host %s.", c.IP)
			} else {
				loger.Log.Errorf("Passive-recvCommand: Failed to receive data sent from CommandSocket from host %s", c.IP)
			}
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
func (c *CommandProcess) recvData(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	go func() {
		select {
		case <-c.ctxProcess.Done():
			loger.Log.Debugf("recvCommand cancel: %s", c.ctxProcess.Err())
			wg.Done()
			return
		}
	}()
	for {
		lengthBuf := make([]byte, 2)
		_, err := io.ReadFull(c.DataSocket, lengthBuf)
		if err != nil {
			if err == io.EOF {
				loger.Log.Debugf("Passive-recvData: Passive connection disconnected from host %s.", c.IP)

				return
			} else {
				continue
			}
		}

		buf := make([]byte, binary.BigEndian.Uint16(lengthBuf)) // 数据接收切片
		n, err := c.DataSocket.Read(buf)
		if err != nil {
			if err == io.EOF {
				loger.Log.Debugf("Passive-recvData: Passive connection disconnected from host %s.", c.IP)
				return
			} else {
				continue
			}
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
