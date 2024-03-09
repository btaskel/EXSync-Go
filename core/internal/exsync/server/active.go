package server

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/exsync/client"
	"EXSync/core/internal/modules/lan"
	loger "EXSync/core/log"
	"EXSync/core/option/exsync/manage"
	serverOption "EXSync/core/option/exsync/server"
	"strings"
	"sync"
	"time"
)

var VerifyManage map[string]serverOption.VerifyManage

// InitClient 主动创建客户端连接对方
// 如果已经预验证，那么直接连接即可通过验证
func (s *Server) InitClient(ip string) {
	c, ok := client.NewClient(ip, s.ActiveConnectManage, VerifyManage)
	if ok {
		s.ActiveConnectManage[ip] = manage.ActiveConnectManage{
			ID:         c.RemoteID,
			CreateTime: time.Now(),
			Client:     c,
		}
	} else {
		c.Close()
	}
}

// GetDevices 运行设备扫描
// LAN模式：逐一扫描局域网并自动搜寻具有正确密钥的计算机
// 白名单模式：在此模式下只有添加的ip才能连接
// 黑名单模式：在此模式下被添加的ip将无法连接
func (s *Server) getDevices() {
	ipSet := make(map[string]struct{})

	scan := config.Config.Server.Scan
	switch strings.ToLower(scan.Type) {
	case "lan":
		loger.Log.Debug("scan: LAN Searching for IP is starting")
		devices, err := lan.ScanDevices()
		if err != nil {
			return
		}
		loger.Log.Debug("scan: LAN Search for IP completed")
		for _, device := range devices {
			ipSet[device] = struct{}{}
		}
	case "white":
		loger.Log.Debug("scan: White List. Search for IP completed")
		for _, device := range scan.Devices {
			ipSet[device] = struct{}{}
		}
	case "black":
		loger.Log.Debug("scan: Black List. Search for IP completed")
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
func (s *Server) checkDevices(ipSet map[string]struct{}) {
	// 如果没有扫描到已经验证的ip, 则删除在VerifyManage中的ip
	for verifyIp := range VerifyManage {
		_, ok := ipSet[verifyIp]
		if !ok {
			delete(VerifyManage, verifyIp)
		}
	}

	// 如果扫描到的Ip已经在VerifyManage中存在，则删除在ipSet中的ip
	for NewIps := range ipSet {
		_, ok := VerifyManage[NewIps]
		if ok {
			delete(ipSet, NewIps)
		}
	}

	wait := sync.WaitGroup{}
	wait.Add(len(ipSet))
	for ip := range ipSet {
		addr := ip
		go func() {
			s.InitClient(addr)
			wait.Done()
		}()
	}
	wait.Wait()
	return
}
