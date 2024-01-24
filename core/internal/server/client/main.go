package client

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/socket"
	"EXSync/core/internal/modules/timechannel"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"time"
)

type Client struct {
	AesGCM        *encryption.Gcm
	CommandSocket net.Conn
	TimeChannel   timechannel.TimeChannel
	ClientMark    string
	DataSocket    net.Conn
	IP            string
	Id            string
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

func (c *Client) connectRemoteCommandSocket() (error error) {
	connectVerify := func(debugStatus bool) bool {
		session := socket.NewSession()
		// 4.本地发送sha384:发送本地密码sha384
		passwordSha384 := hashext.GetSha384(config.Config.Server.Addr.Password)
		base64EncryptLocalID, err := c.AesGCM.StrB64GCMEncrypt(c.Id)
		if err != nil {
			return false
		}
		command := map[string]map[string]string{
			"data": {
				"password_hash": passwordSha384,
				"id":            base64EncryptLocalID,
			},
		}
		result :=
	}
	connectVerifyNoPassword := func(publicKey string, output bool) bool {

	}

	direct := func() bool {
		address := c.IP + ":" + strconv.Itoa(c.CommandSocketPort)
		conn, err := net.DialTimeout("tcp", address, time.Duration(4)*time.Second)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				error = errors.New("timeout")
				return false
			} else {
				error = errors.New("unknownError")
			}
		}
		c.CommandSocket = conn
	}
	check := func() bool {
		for i := 0; i < 3; i++ {
			logrus.Debugf("Connecting to server %v for the %vth time", c.IP, i)
			address := c.IP + ":" + strconv.Itoa(c.CommandSocketPort)
			conn, err := net.DialTimeout("tcp", address, time.Duration(4)*time.Second)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					error = errors.New("timeout")
					return false
				} else {
					error = errors.New("unknownError")
				}
			}
			// 1.本地发送验证指令:发送指令开始进行验证
			command := map[string]interface{}{
				"command": "comm",
				"type":    "verifyconnect",
				"method":  "post",
				"data": map[string]string{
					"version": "0.01",
				},
			}
			jsonData, err := json.Marshal(command)
			if err != nil {
				return false
			}

		}

	}

}
