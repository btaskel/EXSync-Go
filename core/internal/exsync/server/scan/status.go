package scan

import (
	"EXSync/core/internal/config"
	serverOption "EXSync/core/option/exsync/server"
	"github.com/sirupsen/logrus"
)

// success 验证成功
func success(ip, decryptRemoteID string) {
	logrus.Debugf("%s verified", ip)
	VerifyManage[ip] = serverOption.VerifyManage{
		AesKey:   config.Config.Server.Addr.Password,
		RemoteID: decryptRemoteID,
		Permissions: map[string]struct{}{
			"r": {},
		},
	}
}
