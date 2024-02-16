package base

import (
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	serverOption "EXSync/core/option/exsync/server"
	"net"
)

type Base struct {
	Ip                        string
	TimeChannel               *timechannel.TimeChannel
	DataSocket, CommandSocket net.Conn
	AesGCM                    *encryption.Gcm
	VerifyManage              map[string]serverOption.VerifyManage

	//block          uint
	EncryptionLoss int
}
