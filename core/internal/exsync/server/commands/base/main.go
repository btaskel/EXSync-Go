package base

import (
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	"net"
)

type CommandSet struct {
	Ip                        string
	TimeChannel               *timechannel.TimeChannel
	DataSocket, CommandSocket net.Conn
	AesGCM                    *encryption.Gcm
}
