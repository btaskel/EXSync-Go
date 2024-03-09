package base

import (
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	serverOption "EXSync/core/option/exsync/server"
	"EXSync/core/option/exsync/trans"
	"context"
	"net"
)

type Base struct {
	Ip                        string
	TimeChannel               *timechannel.TimeChannel
	DataSocket, CommandSocket net.Conn
	AesGCM                    *encryption.Gcm
	VerifyManage              map[string]serverOption.VerifyManage

	EncryptionLoss int
	CtxProcess     context.Context
	TaskManage     map[string]trans.TranTask
}

// CheckPermission 检测当前操作与验证信息里的权限是否匹配
func CheckPermission(verifyManage serverOption.VerifyManage, permissions []string) bool {
	for _, permission := range permissions {
		if _, ok := verifyManage.Permissions[permission]; !ok {
			return false
		}
	}
	return true
}
