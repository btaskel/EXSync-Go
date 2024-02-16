package server

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/exsync/server/commands/ext"
	serverOption "EXSync/core/option/exsync/server"
	"net"
	"runtime"
	"time"
)

type Server struct {
	ActiveConnectManage  map[string]serverOption.ActiveConnectManage  // 当前主机主动连接远程主机的实例管理
	PassiveConnectManage map[string]serverOption.PassiveConnectManage // 当前主机被动连接远程主机的实例管理
	StopNewConnections   bool
	mergeSocketDict      map[string]map[string]net.Conn
	commandSet           *ext.CommandSet
}

func NewServer() *Server {
	// 设置使用线程
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 创建服务实例
	server := Server{
		ActiveConnectManage:  make(map[string]serverOption.ActiveConnectManage),
		PassiveConnectManage: make(map[string]serverOption.PassiveConnectManage),
		mergeSocketDict:      make(map[string]map[string]net.Conn),
	}

	// 创建局域网扫描验证服务
	go func() {
		for {
			if server.StopNewConnections {
				return
			}
			server.getDevices()
			time.Sleep(10 * time.Second)
		}
	}()

	// 创建监听套接字
	port := config.Config.Server.Addr.Port
	go server.createDataSocket(port)
	go server.createCommandSocket(port + 1)

	return &server
}
