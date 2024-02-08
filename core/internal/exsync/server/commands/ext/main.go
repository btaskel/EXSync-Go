package ext

import (
	"EXSync/core/internal/exsync/server/commands/base"
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	"EXSync/core/option"
	"EXSync/core/option/server/comm"
	"net"
)

type CommandSet struct {
	base.Base
}

// NewCommandSet 创建扩展指令集对象
func NewCommandSet(timeChannel *timechannel.TimeChannel, dataSocket, commandSocket net.Conn, AesGCM *encryption.Gcm,
	verifyManage *map[string]option.VerifyManage, EncryptionLoss int) (commandSet *CommandSet, err error) {
	addr := commandSocket.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	c := CommandSet{base.Base{
		Ip:             ip,
		TimeChannel:    timeChannel,
		DataSocket:     dataSocket,
		CommandSocket:  commandSocket,
		AesGCM:         AesGCM,
		VerifyManage:   verifyManage,
		EncryptionLoss: EncryptionLoss,
	}}

	return &c, err
}

// MatchCommand 匹配命令到相应的函数
func (c *CommandSet) MatchCommand(command comm.Command) {
	switch command.Command {
	case "comm":
		switch command.Type {
		case "verifyConnect":
		case "command":
		case "shell":
		}
	case "data":
		switch command.Type {
		case "file":
			switch command.Method {
			case "get":
				c.GetFile(command.Data)
			case "post":
			}
		case "folder":
			switch command.Method {
			case "get":
			case "post":
			}
		}
	}
}
