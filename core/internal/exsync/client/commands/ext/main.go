package ext

import (
	"EXSync/core/internal/exsync/client/commands/base"
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	"net"
)

type CommandSet struct {
	base.Base
}

func NewCommandSet(ip string, dataSocket, commandSocket net.Conn, timeChannel *timechannel.TimeChannel, gcm *encryption.Gcm, EncryptionLoss int) *CommandSet {
	return &CommandSet{base.Base{
		Ip:             ip,
		TimeChannel:    timeChannel,
		DataSocket:     dataSocket,
		CommandSocket:  commandSocket,
		AesGCM:         gcm,
		EncryptionLoss: EncryptionLoss,
	}}
}
