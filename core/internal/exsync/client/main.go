package client

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/exsync/client/commands/ext"
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/socket"
	"EXSync/core/internal/modules/timechannel"
	"EXSync/core/internal/proxy"
	"EXSync/core/option"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

// Client
// ClientMark 当前客户端实例的标识
// ID 连接对方时表示本机是同一用户的标识
// RemoteID 连接对方时对方表示自己是同一用户的表示
// IP 对方ip地址
type Client struct {
	ext.CommandSet
	AesGCM                       *encryption.Gcm
	TimeChannel                  *timechannel.TimeChannel
	ClientMark, IP, ID, RemoteID string
	commandSocket, dataSocket    net.Conn
}

func NewClient(clientMark, ip string, timeChannel *timechannel.TimeChannel) (client Client, ok bool) {
	gcm, err := encryption.NewGCM(config.Config.Server.Addr.Password)
	if err != nil {
		logrus.Errorf("NewClient: Failed to create AES-GCM! %s", ip)
		return Client{}, false
	}

	client = Client{
		ClientMark:  clientMark,
		IP:          ip,
		ID:          config.Config.Server.Addr.ID,
		AesGCM:      gcm,
		TimeChannel: timeChannel,
	}
	err = client.initSocket()
	if err != nil {
		logrus.Errorf("NewClient: Failed to initialize socket! %s", ip)
		return Client{}, false
	}
	ok = client.connectRemoteCommandSocket()
	if ok {
		return client, true
	}
	logrus.Warningf("NewClient: Verification of connection identity with %s failed", ip)
	return Client{}, false
}

func (c *Client) initSocket() (err error) {
	addr := fmt.Sprintf("%s:%d", c.IP, config.Config.Server.Addr.Port+1)
	if config.Config.Server.Proxy.Enabled {
		c.commandSocket, err = proxy.Socks5.Dial("tcp", addr)
	} else {
		c.commandSocket, err = net.DialTimeout("tcp", addr, 4*time.Second)
	}
	if err != nil {
		logrus.Warningf("Client initSocket: Connection to host %s timeout!", c.IP)
		return err
	}
	addr = fmt.Sprintf("%s:%d", c.IP, config.Config.Server.Addr.Port)
	if config.Config.Server.Proxy.Enabled {
		c.dataSocket, err = proxy.Socks5.Dial("tcp", addr)
	} else {
		c.dataSocket, err = net.DialTimeout("tcp", addr, 4*time.Second)
	}
	if err != nil {
		logrus.Warningf("Client initSocket: Connection to host %s timeout!", c.IP)
		return err
	}
	return nil
}

func (c *Client) connectRemoteCommandSocket() (ok bool) {
	//sessionMark := hashext.GetRandomStr(8)
	//session, err := socket.NewSession(c.TimeChannel, c.DataSocket, c.CommandSocket, sessionMark, "")

	connectVerify := func() bool {

		// 4.本地发送sha384:发送本地密码sha384
		passwordSha384 := hashext.GetSha384(config.Config.Server.Addr.Password)
		base64EncryptLocalID, err := c.AesGCM.B64GCMEncrypt([]byte(c.ID))
		if err != nil {
			logrus.Debug("connectVerify: Encryption base64EncryptLocalID failed!")
			return false
		}
		command := option.Command{
			Command: "",
			Type:    "",
			Method:  "",
			Data: map[string]any{
				"password_hash": passwordSha384,
				"id":            base64EncryptLocalID,
			},
		}
		result, err := socket.SendCommandNoTimeDict(c.dataSocket, command, true)
		if err != nil {
			logrus.Debugf("connectVerify: Sending data failed! %s", err)
			return false
		}
		// 6.远程发送状态和id:获取通过状态和远程id 验证结束
		data, ok := result["data"].(map[string]any)
		if !ok {
			logrus.Debug("connectVerify: Verifying remote connection missing parameter data")
			return false
		}
		encryptRemoteID, ok := data["ID"].(string)
		if !ok {
			logrus.Debug("connectVerify: Verifying remote connection missing parameter ID")
			return false
		}

		remoteID, err := c.AesGCM.B64GCMDecrypt(encryptRemoteID)
		if err != nil {
			logrus.Debugf("connectVerify: B64GCMDecrypt encryptRemoteID failed! %s", err)
			return false
		}

		status, ok := data["status"].(string)
		switch status {
		case "success":
			c.AesGCM, err = encryption.NewGCM(config.Config.Server.Addr.Password)
			if err != nil {
				logrus.Debugf("connectVerify: NewGCM Password failed! %s", err)
				return false
			}
			c.RemoteID = string(remoteID)
			return true
		case "fail":
			logrus.Errorf("connectVerify: Failed to verify server %s password!", c.IP)
			return false
		default:
			logrus.Errorf("connectVerify: Unknown parameter obtained while verifying server %s password!", c.IP)
			return false
		}

	}

	connectVerifyNoPassword := func(publicKey []byte) bool {
		//publicKey, privateKey, err := encryption.GenerateKey()
		//if err != nil {
		//	return false
		//}
		// 将 PEM 块解析为公钥
		publicKeyBlock, _ := pem.Decode(publicKey)
		if publicKeyBlock == nil {
			logrus.Debug("connectVerifyNoPassword: failed to decode PEM block containing public key!")
			return false
		}
		publicKeyInterface, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
		if err != nil {
			logrus.Debugf("connectVerifyNoPassword: failed to parse RSA public key: %v", err)
			return false
		}
		// 将公钥转换为 *rsa.PublicKey
		aesPublicKey, ok := publicKeyInterface.(*rsa.PublicKey)
		if !ok {
			logrus.Debug("connectVerifyNoPassword: not an RSA public key")
			return false
		}
		sessionPassword := hashext.GetRandomStr(16)
		base64EncryptSessionPassword, err := encryption.RsaEncryptBase64([]byte(sessionPassword), aesPublicKey)
		if err != nil {
			logrus.Debugf("connectVerifyNoPassword: Encrypting base64EncryptSessionPassword with publicKey %s failed!", publicKey)
			return false
		}
		base64EncryptSessionID, err := encryption.RsaEncryptBase64([]byte(c.ID), aesPublicKey)
		if err != nil {
			logrus.Debugf("connectVerifyNoPassword: Encrypting base64EncryptSessionID with publicKey %s failed!", publicKey)
			return false
		}
		replyCommand := option.Command{
			Command: "",
			Type:    "",
			Method:  "",
			Data: map[string]any{
				"session_password": base64EncryptSessionPassword,
				"id":               base64EncryptSessionID,
			},
		}
		result, err := socket.SendCommandNoTimeDict(c.dataSocket, replyCommand, true)
		if err != nil {
			logrus.Debugf("connectVerifyNoPassword: Sending data failed! %s", err)
			return false
		}
		data, ok := result["data"].(map[string]any)
		if !ok {
			logrus.Debug("connectVerifyNoPassword: Verifying remote connection missing parameter data")
			return false
		}
		remoteID, ok := data["id"].(string)
		if !ok {
			logrus.Debug("connectVerifyNoPassword: Verifying remote connection missing parameter remoteID")
			return false
		}
		gcm, err := encryption.NewGCM(sessionPassword)
		if err != nil {
			logrus.Debug("connectVerifyNoPassword: NewGCM: Failed to create Cipher with key!")
			return false
		}
		remoteID, err = gcm.StrB64GCMDecrypt(remoteID)
		if err != nil {
			logrus.Debug("connectVerifyNoPassword: Encryption remoteID failed!")
			return false
		}
		c.RemoteID = remoteID
		c.AesGCM = gcm
		return true
	}

	direct := func() bool {
		address := c.IP + ":"
		_, err := net.DialTimeout("tcp", address, time.Duration(4)*time.Second)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logrus.Debugf("direct: Connection to %s failed", c.IP)
				//error = errors.New("timeout")
				return false
			} else {
				//error = errors.New("unknownError")
				return false
			}
		}
		return true
		//c.CommandSocket = conn
	}
	check := func() bool {
		for i := 0; i < 3; i++ {
			// 1.本地发送验证指令:发送指令开始进行验证
			logrus.Debugf("check: Connecting to server %s for the %vth time", c.IP, i)
			command := option.Command{
				Command: "comm",
				Type:    "verifyConnect",
				Method:  "post",
				Data: map[string]any{
					"version": config.Config.Version,
				},
			}
			result, err := socket.SendCommandNoTimeDict(c.commandSocket, command, true)
			if err != nil {
				return false
			}
			data, ok := result["data"].(map[string]any)
			if !ok {
				return false
			}
			// 3.远程发送sha256值:验证远程sha256值是否与本地匹配
			publicKey, publicKeyOk := data["public_key"].(string)
			remotePasswordSha256, remotePasswordSha256Ok := data["password_hash"].(string)
			if remotePasswordSha256Ok && remotePasswordSha256 == hashext.GetSha256(config.Config.Server.Addr.Password) {
				if connectVerify() {
					return true
				} else {
					return false
				}
			} else if publicKeyOk && !remotePasswordSha256Ok {
				logrus.Infof("check: Target server %s has no password set.", c.IP)
				if connectVerifyNoPassword([]byte(publicKey)) {
					return true
				} else {
					return false
				}
			}
		}
		logrus.Debugf("check: Verification failed with host X")
		return false
	}
	if c.AesGCM != nil {
		return direct()
	} else {
		return check()
	}

}
