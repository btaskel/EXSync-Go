package base

import (
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	"EXSync/core/option"
	"net"
)

type Base struct {
	Ip                        string
	TimeChannel               *timechannel.TimeChannel
	DataSocket, CommandSocket net.Conn
	AesGCM                    *encryption.Gcm
	VerifyManage              *map[string]option.VerifyManage

	//block          uint
	EncryptionLoss int
}
