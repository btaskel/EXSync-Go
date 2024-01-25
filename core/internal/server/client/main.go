package client

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/socket"
	"EXSync/core/internal/modules/timechannel"
	"EXSync/core/option"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type Client struct {
	AesGCM        *encryption.Gcm
	CommandSocket net.Conn
	TimeChannel   *timechannel.TimeChannel
	ClientMark    string
	DataSocket    net.Conn
	IP            string
	Id            string
	RemoteID      string
}

func NewClient(commandSocket, dataSocket net.Conn, ip, clientMark, id string, AesGCM *encryption.Gcm) *Client {
	return &Client{
		AesGCM:        AesGCM,
		ClientMark:    clientMark,
		CommandSocket: commandSocket,
		DataSocket:    dataSocket,
		IP:            ip,
		Id:            id,
	}
}

func (c *Client) setProxy() {
	return
}

func (c *Client) initSocket() (error error) {

	return
}

func (c *Client) connectRemoteCommandSocket() (ok bool) {
	//sessionMark := hashext.GetRandomStr(8)
	//session, err := socket.NewSession(c.TimeChannel, c.DataSocket, c.CommandSocket, sessionMark, "")

	connectVerify := func() bool {

		// 4.本地发送sha384:发送本地密码sha384
		passwordSha384 := hashext.GetSha384(config.Config.Server.Addr.Password)
		base64EncryptLocalID, err := c.AesGCM.B64GCMEncrypt([]byte(c.Id))
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
		result, err := socket.SendCommandNoTimeDict(c.DataSocket, command, true)
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
			logrus.Errorf("Failed to verify server %s password!", c.IP)
			return false
		default:
			logrus.Errorf("Unknown parameter obtained while verifying server %s password!", c.IP)
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
		base64EncryptSessionID, err := encryption.RsaEncryptBase64([]byte(c.Id), aesPublicKey)
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
		result, err := socket.SendCommandNoTimeDict(c.DataSocket, replyCommand, true)
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
			return false
		}
		gcm, err := encryption.NewGCM(sessionPassword)
		if err != nil {
			return false
		}
		remoteID, err = gcm.StrB64GCMDecrypt(remoteID)
		if err != nil {
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
			logrus.Debugf("Connecting to server %s for the %vth time", c.IP, i)
			command := option.Command{
				Command: "comm",
				Type:    "verifyConnect",
				Method:  "post",
				Data: map[string]any{
					"version": config.Config.Version,
				},
			}
			result, err := socket.SendCommandNoTimeDict(c.CommandSocket, command, true)
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
				logrus.Infof("Target server %s has no password set.", c.IP)
				if connectVerifyNoPassword([]byte(publicKey)) {
					return true
				} else {
					return false
				}
			}
		}
		logrus.Debugf("Verification failed with host X")
		return false
	}
	if c.AesGCM != nil {
		return direct()
	} else {
		return check()
	}

}
