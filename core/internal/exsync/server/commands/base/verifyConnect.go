package base

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/socket"
	"EXSync/core/option"
	"github.com/sirupsen/logrus"
)

func (c *CommandSet) VerifyConnect(data map[string]any, mark string) {
	defer c.TimeChannel.DelKey(mark) // 释放当前会话
	session, err := socket.NewSession(c.TimeChannel, c.DataSocket, nil, mark, "")
	if err != nil {
		return
	}
	version, ok := data["version"].(float64)
	if !ok {
		logrus.Errorf("Host %s : Missing parameter [version] for execution [verifyConnect]", c.Ip)
		return
	}
	if version != config.Config.Version {
		// 客户方与服务端验证版本不一致
		return
	}
	if config.Config.Server.Addr.Password != "" {
		// 如果密码为空, 则与客户端交换AES密钥

	} else {
		// 2.验证远程sha256值是否与本地匹配: 发送本地的password_sha256值到远程
		// 如果密码不为空, 则无需进行密钥交换, 只需验证密钥即可
		passwordSha256 := hashext.GetSha256(config.Config.Server.Addr.Password)
		command := option.Command{
			Command: "",
			Type:    "",
			Method:  "",
			Data: map[string]any{
				"password_hash": passwordSha256,
			},
		}
		result, err := session.SendCommand(command, true, false)
		if err != nil {
			return
		}
		resultData, ok := result["data"].(map[string]any)
		if !ok {
			return
		}

		remotePasswordSha384, ok := resultData["password_hash"].(string)
		if !ok {
			return
		}
		remoteIDEncryptBase64, ok := resultData["id"].(string)
		if !ok {
			return
		}
		gcm, err := encryption.NewGCM(config.Config.Server.Addr.Password)
		if err != nil {
			return
		}
		remoteID, err := gcm.B64GCMDecrypt(remoteIDEncryptBase64)
		if err != nil {
			return
		}
		// 5.验证本地sha384值是否与远程匹配匹配: 接收对方的密码sha384值, 如果通过返回id和验证状态
		localPasswordSha384 := hashext.GetSha384(config.Config.Server.Addr.Password)
		//encryptLocalID, err := gcm.AesGcmEncrypt([]byte(config.Config.Server.Addr.ID))
		//if err != nil {
		//	return
		//}
		localIDEncryptBase64, err := gcm.B64GCMEncrypt([]byte(config.Config.Server.Addr.ID))
		if err != nil {
			return
		}
		if remotePasswordSha384 == localPasswordSha384 {
			command = option.Command{
				Command: "",
				Type:    "",
				Method:  "",
				Data: map[string]any{
					"status": "success",
					"id":     localIDEncryptBase64,
				},
			}
			_, err = session.SendCommand(command, false, true)
			if err != nil {
				return
			}
			// todo:验证成功
		} else {
			command = option.Command{
				Command: "",
				Type:    "",
				Method:  "",
				Data: map[string]any{
					"status": "fail",
				},
			}
			_, err = session.SendCommand(command, false, true)
		}
	}
	return
}
