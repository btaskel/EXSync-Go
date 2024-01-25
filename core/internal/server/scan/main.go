package scan

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/socket"
	"EXSync/core/internal/server/scan/lan"
	"EXSync/core/option"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
	"time"
)

type Scan struct {
	VerifyManage map[string]option.VerifyManage
}

//// NewScan 创建一个设备扫描实例
//func NewScan() *Scan {
//	return &Scan{
//		verifyManage: nil,
//	}
//}

// ScanDevices 运行设备扫描
// LAN模式：逐一扫描局域网并自动搜寻具有正确密钥的计算机
// 白名单模式：在此模式下只有添加的ip才能连接
// 黑名单模式：在此模式下被添加的ip将无法连接
func (s *Scan) ScanDevices() {
	ipSet := make(map[string]struct{})

	scan := config.Config.Server.Scan
	switch strings.ToLower(scan.Type) {
	case "lan":
		logrus.Debug("scan: LAN Searching for IP is starting")
		devices, err := lan.ScanDevices()
		if err != nil {
			return
		}
		logrus.Debug("scan: LAN Search for IP completed")
		for _, device := range devices {
			ipSet[device] = struct{}{}
		}
		//ipSet = append(ipSet, devices...)
	case "white":
		logrus.Debug("scan: White List. Search for IP completed")
		for _, device := range scan.Devices {
			ipSet[device] = struct{}{}
		}
	case "black":
		logrus.Debug("scan: Black List. Search for IP completed")
		for _, device := range scan.Devices {
			ipSet[device] = struct{}{}
		}
	}

	blockIpList := []string{
		"192.168.1.1",
	}

	for _, ip := range blockIpList {
		delete(ipSet, ip)
	}

	s.checkDevices(ipSet) // 将会更新verifyManage
}

// checkDevices 检查设备并返回当前批次已验证列表
// 主动验证：主动嗅探并验证ip列表是否存在活动的设备, 如果存在活动的设备判断密码是否相同
// [1,4] [2,3]
func (s *Scan) checkDevices(ipSet map[string]struct{}) {
	wait := sync.WaitGroup{}
	for ip := range s.VerifyManage {
		_, ok := ipSet[ip]
		if !ok {
			delete(s.VerifyManage, ip)
		}
	}
	for NewIps := range ipSet {
		_, ok := s.VerifyManage[NewIps]
		if ok {
			delete(ipSet, NewIps)
		}
	}
	wait.Add(len(ipSet))
	for NewIps := range ipSet {
		go s.connectServer(NewIps, &wait) // 判断主机是否通过验证
	}
	wait.Wait()
	return
}

// connectServer 连接需验证方commandSocket并进行验证
func (s *Scan) connectServer(ip string, wait *sync.WaitGroup) {
	defer wait.Done()
	// 连接指定端口
	address := fmt.Sprintf("%s:%d", ip, config.Config.Server.Addr.Port+1)
	conn, err := net.DialTimeout("tcp", address, time.Duration(4)*time.Second)
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		err = conn.Close()
		if err != nil {
			return
		}
		return
	}
	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(conn)
	command := option.Command{
		Command: "comm",
		Type:    "verifyConnect",
		Method:  "post",
		Data: map[string]any{
			"version": config.Config.Version,
		},
	}
	gcm, err := encryption.NewGCM(config.Config.Server.Addr.Password)
	if err != nil {
		return
	}
	{
		mark := hashext.GetRandomStr(8)
		session, err := socket.NewSession(nil, nil, conn, mark, "")
		if err != nil {
			logrus.Debugf("scan : An error occurred while establishing a new session for %s", ip)
			return
		}
		// 3.远程发送sha256值:验证远程sha256值是否与本地匹配
		result, err := session.SendCommand(command, true, true)
		if err != nil {
			logrus.Debugf("scan : An error occurred while verifying whether the remote sha256 value matches the local value, from %s", ip)
			return
		}

		remotePasswordSha256, remotePasswordSha256ok := result["password_hash"].(string)
		publicKey, publicKeyOk := result["public_key"].(string)

		if remotePasswordSha256 == hashext.GetSha256(config.Config.Server.Addr.Password) {
			// 4.本地发送sha384:发送本地密码sha384
			localPasswordSha384 := hashext.GetSha384(config.Config.Server.Addr.Password)
			base64EncryptLocalID, err := gcm.B64GCMEncrypt([]byte(config.Config.Server.Addr.ID))
			if err != nil {
				logrus.Debugf("An error occurred while encrypting the local ID using base64 + AES-CTR, when sending to %s", ip)
				return
			}
			//replyCommand := map[string]map[string]string{
			//	"data": {
			//		"password_hash": localPasswordSha384,
			//		"id":            base64EncryptLocalID,
			//	},
			//}
			replyCommand := option.Command{
				Command: "",
				Type:    "",
				Method:  "",
				Data: map[string]any{
					"password_hash": localPasswordSha384,
					"id":            base64EncryptLocalID,
				},
			}
			result, err = session.SendCommand(replyCommand, true, true)
			if err != nil {
				logrus.Debugf("")
				return
			}
			// 6.远程发送状态和id:获取通过状态和远程id 验证结束
			data, ok := result["data"].(map[string]string)
			if !ok {
				return
			}
			status, ok := data["password_hash"]
			if !ok {
				return
			}
			remoteID, ok := data["id"]
			if !ok {
				return
			}
			decryptRemoteID, err := gcm.B64GCMDecrypt(remoteID)
			if err != nil {
				return
			}
			switch status {
			case "success":
				// 验证成功
				s.success(ip, string(decryptRemoteID))
				return
			case "fail":
				// 验证服务端密码失败
				logrus.Infof("scan: Verification failed when connecting to server %v", ip)
				return
			default:
				// 验证服务端时得到未知参数
				return
			}
		} else if !remotePasswordSha256ok && publicKeyOk {
			// 对方密码为空，示意任何设备均可连接
			//publicKey, privateKey, err := encryption.GenerateKey()
			//n := publicKey.N.String()
			//e := strconv.Itoa(publicKey.E)
			// 明文^E%N = 密文
			// 密文^D%N = 明文
			// 将公钥字符串解码为字节切片
			publicKeyBytes := []byte(publicKey)
			publicKeyBlock, _ := pem.Decode(publicKeyBytes)
			if publicKeyBlock == nil {
				logrus.Debug("scan: failed to decode PEM block containing public key!")
				return
			}
			// 将 PEM 块解析为公钥
			publicKeyInterface, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
			if err != nil {
				logrus.Debugf("scan: failed to parse RSA public key: %v", err)
				return
			}
			// 将公钥转换为 *rsa.PublicKey
			aesPublicKey, ok := publicKeyInterface.(*rsa.PublicKey)
			if !ok {
				logrus.Debug("scan: not an RSA public key")
				return
			}
			// 加密会话密钥
			sessionPassword := hashext.GetRandomStr(16)
			sessionPasswordEncryptBase64, err := encryption.RsaEncryptBase64([]byte(sessionPassword), aesPublicKey)
			if err != nil {
				return
			}
			// 加密会话id
			sessionIDEncryptBase64, err := encryption.RsaEncryptBase64([]byte(config.Config.Server.Addr.ID), aesPublicKey)
			if err != nil {
				return
			}
			replyCommand := option.Command{
				Command: "",
				Type:    "",
				Method:  "",
				Data: map[string]any{
					"session_password": sessionPasswordEncryptBase64,
					"id":               sessionIDEncryptBase64,
				},
			}
			result, err = session.SendCommand(replyCommand, true, true)
			if err != nil {
				logrus.Debugf("A timeout error occurred while sending encrypted session keys to host %s!", ip)
				return
			}
			data, ok := result["data"].(map[string]any)
			if !ok {
				return
			}
			remoteID, ok := data["id"].(string)
			if !ok {
				return
			}
			remoteOriginID, err := gcm.B64GCMDecrypt(remoteID)
			if err != nil {
				return
			}
			s.success(ip, string(remoteOriginID))
			return

		} else {
			// 验证客户端密码哈希得到未知参数
			return
		}
	}
}
