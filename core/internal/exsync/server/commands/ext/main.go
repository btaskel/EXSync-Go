package ext

import (
	"EXSync/core/internal/exsync/server/commands/base"
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	"EXSync/core/option"
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
