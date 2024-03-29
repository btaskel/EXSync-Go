package client

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/socket"
	loger "EXSync/core/log"
	"EXSync/core/option/exsync/comm"
	serverOption "EXSync/core/option/exsync/server"
	"time"
)

//func (c *Client) connectRemoteCommandSocket() (ok bool) {
//	//sessionMark := hashext.GetRandomStr(8)
//	//session, err := socket.NewSession(c.TimeChannel, c.DataSocket, c.CommandSocket, sessionMark, "")
//
//	connectVerify := func() bool {
//		// 4.本地发送sha384:发送本地密码sha384
//		passwordSha384 := hashext.GetSha384(config.Config.Server.Addr.Password)
//		base64EncryptLocalID, err := c.AesGCM.B64GCMEncrypt([]byte(c.ID))
//		if err != nil {
//			loger.Log.Debug("connectVerify: Encryption base64EncryptLocalID failed!")
//			return false
//		}
//		replyCommand := comm.Command{
//			Command: "",
//			Type:    "",
//			Method:  "",
//			Data: map[string]any{
//				"password_hash": passwordSha384,
//				"id":            base64EncryptLocalID,
//			},
//		}
//		result, err := socket.SendCommandNoTimeDict(c.dataSocket, replyCommand, true)
//		if err != nil {
//			loger.Log.Debugf("connectVerify: Sending data failed! %s", err)
//			return false
//		}
//		// 6.远程发送状态和id:获取通过状态和远程id 验证结束
//		data, ok := result["data"].(map[string]any)
//		if !ok {
//			loger.Log.Debug("connectVerify: Verifying remote connection missing parameter data")
//			return false
//		}
//		encryptRemoteID, ok := data["ID"].(string)
//		if !ok {
//			loger.Log.Debug("connectVerify: Verifying remote connection missing parameter ID")
//			return false
//		}
//
//		remoteID, err := c.AesGCM.B64GCMDecrypt(encryptRemoteID)
//		if err != nil {
//			loger.Log.Debugf("connectVerify: B64GCMDecrypt encryptRemoteID failed! %s", err)
//			return false
//		}
//
//		status, ok := data["status"].(string)
//		switch status {
//		case "success":
//			c.AesGCM, err = encryption.NewGCM(config.Config.Server.Addr.Password)
//			if err != nil {
//				loger.Log.Debugf("connectVerify: NewGCM Password failed! %s", err)
//				return false
//			}
//			c.RemoteID = string(remoteID)
//			return true
//		case "fail":
//			loger.Log.Errorf("connectVerify: Failed to verify server %s password!", c.IP)
//			return false
//		default:
//			loger.Log.Errorf("connectVerify: Unknown parameter obtained while verifying server %s password!", c.IP)
//			return false
//		}
//
//	}
//
//	connectVerifyNoPassword := func(publicKey []byte) bool {
//		//publicKey, privateKey, err := encryption.GenerateKey()
//		//if err != nil {
//		//	return false
//		//}
//		// 将 PEM 块解析为公钥
//		publicKeyBlock, _ := pem.Decode(publicKey)
//		if publicKeyBlock == nil {
//			loger.Log.Debug("connectVerifyNoPassword: failed to decode PEM block containing public key!")
//			return false
//		}
//		publicKeyInterface, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
//		if err != nil {
//			loger.Log.Debugf("connectVerifyNoPassword: failed to parse RSA public key: %v", err)
//			return false
//		}
//		// 将公钥转换为 *rsa.PublicKey
//		aesPublicKey, ok := publicKeyInterface.(*rsa.PublicKey)
//		if !ok {
//			loger.Log.Debug("connectVerifyNoPassword: not an RSA public key")
//			return false
//		}
//		sessionPassword := hashext.GetRandomStr(16)
//		base64EncryptSessionPassword, err := encryption.RsaEncryptBase64([]byte(sessionPassword), aesPublicKey)
//		if err != nil {
//			loger.Log.Debugf("connectVerifyNoPassword: Encrypting base64EncryptSessionPassword with publicKey %s failed!", publicKey)
//			return false
//		}
//		base64EncryptSessionID, err := encryption.RsaEncryptBase64([]byte(c.ID), aesPublicKey)
//		if err != nil {
//			loger.Log.Debugf("connectVerifyNoPassword: Encrypting base64EncryptSessionID with publicKey %s failed!", publicKey)
//			return false
//		}
//		replyCommand := comm.Command{
//			Command: "",
//			Type:    "",
//			Method:  "",
//			Data: map[string]any{
//				"session_password": base64EncryptSessionPassword,
//				"id":               base64EncryptSessionID,
//			},
//		}
//		result, err := socket.SendCommandNoTimeDict(c.dataSocket, replyCommand, true)
//		if err != nil {
//			loger.Log.Debugf("connectVerifyNoPassword: Sending data failed! %s", err)
//			return false
//		}
//		data, ok := result["data"].(map[string]any)
//		if !ok {
//			loger.Log.Debug("connectVerifyNoPassword: Verifying remote connection missing parameter data")
//			return false
//		}
//		remoteID, ok := data["id"].(string)
//		if !ok {
//			loger.Log.Debug("connectVerifyNoPassword: Verifying remote connection missing parameter remoteID")
//			return false
//		}
//		gcm, err := encryption.NewGCM(sessionPassword)
//		if err != nil {
//			loger.Log.Debug("connectVerifyNoPassword: NewGCM: Failed to create Cipher with key!")
//			return false
//		}
//		remoteID, err = gcm.StrB64GCMDecrypt(remoteID)
//		if err != nil {
//			loger.Log.Debug("connectVerifyNoPassword: Encryption remoteID failed!")
//			return false
//		}
//		c.RemoteID = remoteID
//		c.AesGCM = gcm
//		return true
//	}
//
//	direct := func() bool {
//		address := c.IP + ":"
//		_, err := net.DialTimeout("tcp", address, time.Duration(4)*time.Second)
//		if err != nil {
//			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
//				loger.Log.Debugf("direct: Connection to %s failed", c.IP)
//				//error = errors.New("timeout")
//				return false
//			} else {
//				//error = errors.New("unknownError")
//				return false
//			}
//		}
//		return true
//		//c.CommandSocket = conn
//	}
//
//	check := func() bool {
//		for i := 0; i < 3; i++ {
//			// 1.本地发送验证指令:发送指令开始进行验证
//			loger.Log.Debugf("check: Connecting to server %s for the %vth time", c.IP, i)
//			replyCommand := comm.Command{
//				Command: "comm",
//				Type:    "verifyConnect",
//				Method:  "post",
//				Data: map[string]any{
//					"version": config.Config.Version,
//				},
//			}
//			result, err := socket.SendCommandNoTimeDict(c.commandSocket, replyCommand, true)
//			if err != nil {
//				return false
//			}
//			data, ok := result["data"].(map[string]any)
//			if !ok {
//				return false
//			}
//			// 3.远程发送sha256值:验证远程sha256值是否与本地匹配
//			publicKey, publicKeyOk := data["public_key"].(string)
//			remotePasswordSha256, remotePasswordSha256Ok := data["password_hash"].(string)
//			if remotePasswordSha256Ok && remotePasswordSha256 == hashext.GetSha256(config.Config.Server.Addr.Password) {
//				if connectVerify() {
//					return true
//				} else {
//					return false
//				}
//			} else if publicKeyOk && !remotePasswordSha256Ok {
//				loger.Log.Infof("check: Target server %s has no password set.", c.IP)
//				if connectVerifyNoPassword([]byte(publicKey)) {
//					return true
//				} else {
//					return false
//				}
//			}
//		}
//		loger.Log.Debugf("check: Verification failed with host X")
//		return false
//	}
//	if c.AesGCM == nil {
//		return check()
//	} else {
//		return direct()
//	}
//
//}

func (c *Client) connectRemoteCommandSocket() (ok bool, err error) {
	//// 远程密码为空，进行密钥交换
	//noPassword := func() {
	//
	//}

	// 本地密码不为空
	requirePassword := func() bool {
		_, localOffset := time.Now().Zone()
		reply := comm.Command{
			Command: "comm",
			Type:    "verifyConnect",
			Method:  "post",
			Data: map[string]any{
				"version": config.Config.Version, // 当前版本
				"offset":  localOffset / 3600,    // 当前时区
				"id":      c.LocalID,             // 返回当前主机的id标识
				"hash":    hashext.GetSha384(config.Config.Server.Addr.Password),
				//"stat":    "",                    // 对于上个请求的状态反馈
			},
		}
		replyMark := hashext.GetRandomStr(6)
		s, err := socket.NewSession(c.TimeChannel, nil, c.commandSocket, replyMark, c.aesGcm)
		if err != nil {
			return false
		}
		defer s.Close()
		command, err := s.SendCommand(reply, true, true)
		if err != nil {
			return false
		}
		remoteID, ok := command["id"].(string)
		if !ok {
			socket.SendStat(s, "connectRemoteCommandSocket: ")
			return false
		}
		remotePermissions, ok := command["permissions"].(map[string]struct{}) // r 读取; w 写入; e执行;
		if !ok {
			return false
		}
		remoteOffset, ok := command["offset"].(int64)
		if !ok {
			return false
		}

		c.RemoteID = remoteID
		c.VerifyManage[c.IP] = serverOption.VerifyManage{
			AesKey:      config.Config.Server.Addr.Password,
			RemoteID:    remoteID,
			Offset:      remoteOffset*3600 - int64(localOffset),
			Permissions: remotePermissions,
		}
		return true
	}

	// 根据当前密码状态选择验证方式
	if config.Config.Server.Addr.Password == "" {
		// 无密码验证
		return
	} else {
		// 有密码验证
		if requirePassword() {
			loger.Log.Infof("Active: Local password verification with remote host %s was successful.", c.IP)
			return true, nil
		} else {
			loger.Log.Infof("Active: Local password verification with remote host %s failed.", c.IP)
			return false, nil
		}
	}
}
