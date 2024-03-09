package client

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/exsync/client/commands/base"
	"EXSync/core/internal/exsync/client/commands/ext"
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	"EXSync/core/internal/proxy"
	loger "EXSync/core/log"
	"EXSync/core/option/exsync/manage"
	serverOption "EXSync/core/option/exsync/server"
	"fmt"
	"net"
	"time"
)

// Client
// ClientMark 当前客户端实例的标识
// ID 连接对方时表示本机是同一用户的标识
// RemoteID 连接对方时对方表示自己是同一用户
// IP 对方ip地址
type Client struct {
	TimeChannel               *timechannel.TimeChannel
	IP, LocalID, RemoteID     string
	Comm                      *ext.CommandSet
	ActiveConnectManage       map[string]manage.ActiveConnectManage
	VerifyManage              map[string]serverOption.VerifyManage
	aesGcm                    *encryption.Gcm
	commandSocket, dataSocket net.Conn
}

func NewClient(ip string, activeConnectManage map[string]manage.ActiveConnectManage, verifyManage map[string]serverOption.VerifyManage) (*Client, bool) {
	// 初始化AES-GCM
	var gcm *encryption.Gcm
	var err error
	if len(config.Config.Server.Addr.Password) != 0 {
		gcm, err = encryption.NewGCM(config.Config.Server.Addr.Password)
		if err != nil {
			loger.Log.Errorf("NewClient: Failed to create AES-GCM! %s", ip)
			return nil, false
		}
	}

	// 初始化独立的TimeChannel
	timeChannel := timechannel.NewTimeChannel()

	// 初始化Client实例
	client := &Client{
		IP:                  ip,
		LocalID:             config.Config.Server.Addr.ID,
		TimeChannel:         timeChannel,
		ActiveConnectManage: activeConnectManage,
		VerifyManage:        verifyManage,
		aesGcm:              gcm,
	}

	// 初始化Socket, 如果使用socks5代理，则使用socks进行初始化
	err = client.initSocket()
	if err != nil {
		loger.Log.Errorf("NewClient: Failed to initialize socket! %s", ip)
		return nil, false
	}

	// 开始验证并连接远程主机
	ok, err := client.connectRemoteCommandSocket()
	if !ok {
		return nil, false
	}
	if err == nil {
		commBase := base.Base{
			Ip:             ip,
			TimeChannel:    timeChannel,
			DataSocket:     client.dataSocket,
			CommandSocket:  client.commandSocket,
			AesGCM:         gcm,
			VerifyManage:   verifyManage[ip],
			EncryptionLoss: 28,
		}
		client.Comm = ext.NewCommandSet(commBase)
		return client, true
	} else {
		loger.Log.Errorf("NewClient: Verification of connection identity with %s failed! %s", ip, err)
		return nil, false
	}
}

func (c *Client) initSocket() (err error) {
	addr := fmt.Sprintf("%s:%d", c.IP, config.Config.Server.Addr.Port+1)
	if config.Config.Server.Proxy.Enabled {
		c.commandSocket, err = proxy.Socks5.Dial("tcp", addr)
	} else {
		c.commandSocket, err = net.DialTimeout("tcp", addr, 4*time.Second)
	}
	if err != nil {
		loger.Log.Warningf("Client initSocket: Connection to host %s timeout!", c.IP)
		return err
	}
	addr = fmt.Sprintf("%s:%d", c.IP, config.Config.Server.Addr.Port)
	if config.Config.Server.Proxy.Enabled {
		c.dataSocket, err = proxy.Socks5.Dial("tcp", addr)
	} else {
		c.dataSocket, err = net.DialTimeout("tcp", addr, 4*time.Second)
	}
	if err != nil {
		loger.Log.Warningf("Client initSocket: Connection to host %s timeout!", c.IP)
		return err
	}
	return nil
}

// Close 关闭客户端的连接
func (c *Client) Close() bool {
	err := c.commandSocket.Close()
	if err != nil {
		netErr, ok := err.(*net.OpError)
		if ok && netErr.Err.Error() != "use of closed network connection" {
			loger.Log.Errorf("Attempt to close active connection with host %s failed! %s", c.IP, err)
			return false
		}
	}
	err = c.dataSocket.Close()
	if err != nil {
		netErr, ok := err.(*net.OpError)
		if ok && netErr.Err.Error() != "use of closed network connection" {
			loger.Log.Errorf("Attempt to close active connection with host %s failed! %s", c.IP, err)
			return false
		}
	}
	c.TimeChannel.Close()
	delete(c.ActiveConnectManage, c.IP)
	return true
}
