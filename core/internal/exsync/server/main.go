package server

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/exsync/server/commands/ext"
	serverOption "EXSync/core/option/exsync/server"
	"net"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	ActiveConnectManage    map[string]serverOption.ActiveConnectManage  // 当前主机主动连接远程主机的实例管理
	PassiveConnectManage   map[string]serverOption.PassiveConnectManage // 当前主机被动连接远程主机的实例管理
	mergeSocketDict        map[string]map[string]net.Conn
	commandSet             *ext.CommandSet
	commListen, dataListen net.Listener
	stopServer             bool
}

// NewServer 创建传输服务对象
func NewServer() *Server {
	// 设置使用线程
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 创建服务实例
	server := Server{
		ActiveConnectManage:  make(map[string]serverOption.ActiveConnectManage),
		PassiveConnectManage: make(map[string]serverOption.PassiveConnectManage),
		mergeSocketDict:      make(map[string]map[string]net.Conn),
	}

	return &server
}

// Run 运行服务
func (s *Server) Run() {
	// 创建局域网扫描验证服务
	go func() {
		for {
			if s.stopServer {
				return
			}
			s.getDevices()
			time.Sleep(10 * time.Second)
		}
	}()
	go func() {
		// 创建监听套接字
		port := config.Config.Server.Addr.Port

		wait := sync.WaitGroup{}
		wait.Add(2)
		go func() {
			s.createDataSocket(port)
			wait.Done()
		}()
		go func() {
			s.createCommandSocket(port + 1)
			wait.Done()
		}()
		wait.Wait()
	}()
}

// Close 关闭服务
func (s *Server) Close() {
	s.commListen.Close()
	s.dataListen.Close()
	s.stopServer = true
}
