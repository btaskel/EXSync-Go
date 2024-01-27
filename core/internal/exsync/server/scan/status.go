package scan

import (
	"EXSync/core/internal/config"
	"EXSync/core/option"
	"github.com/sirupsen/logrus"
)

// success 验证成功
func (s *Scan) success(ip, decryptRemoteID string) {
	logrus.Debugf("%s verified", ip)
	s.VerifyManage[ip] = option.VerifyManage{
		AesKey:     config.Config.Server.Addr.Password,
		RemoteID:   decryptRemoteID,
		Permission: 0,
	}
}

func (s *Scan) fail() {

}
