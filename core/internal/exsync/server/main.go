package server

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/exsync/client"
	"EXSync/core/internal/exsync/server/commands"
	"EXSync/core/internal/exsync/server/commands/ext"
	"EXSync/core/internal/exsync/server/scan"
	serverOption "EXSync/core/option/exsync/server"
	"fmt"
	"github.com/sirupsen/logrus"
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
			scan.ScanDevices()
			time.Sleep(10 * time.Second)
		}
	}()

	// 创建监听套接字
	port := config.Config.Server.Addr.Port
	go server.createDataSocket(port)
	go server.createCommandSocket(port + 1)

	return &server
}

// createCommandSocket 创建套接字对象
func (s *Server) createCommandSocket(port int) {
	address := fmt.Sprintf("%s:%d", config.Config.Server.Addr.IP, port)
	conn, err := net.Listen("tcp", address)
	if err != nil {
		return
	}
	for {
		if s.StopNewConnections {
			return
		}
		socket, err := conn.Accept()
		if err != nil {
			logrus.Debugf("address %s: %s", address, err)
			continue
		}
		go s.verifyCommandSocket(socket)
	}
}

// createDataSocket 创建套接字对象
func (s *Server) createDataSocket(port int) {
	address := fmt.Sprintf("%s:%d", config.Config.Server.Addr.IP, port)
	conn, err := net.Listen("tcp", address)
	if err != nil {
		return
	}
	for {
		if s.StopNewConnections {
			return
		}
		socket, err := conn.Accept()
		if err != nil {
			logrus.Debugf("address %s: %s", address, err)
			continue
		}
		go s.verifyDataSocket(socket)
	}
}

// verifyCommandSocket 判断对方是否已经通过预扫描验证
func (s *Server) verifyCommandSocket(commandSocket net.Conn) {
	defer func(commandSocket net.Conn) {
		err := commandSocket.Close()
		if err != nil {
			logrus.Warning(err)
		}
	}(commandSocket)
	addr := commandSocket.RemoteAddr().String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	logrus.Infof("Starting to verify command socket connection from %s...", host)
	if hostInfo, ok := scan.VerifyManage[host]; ok && hostInfo.AesKey != "" {
		// todo: 验证通过处理
		if dataSocket, ok := s.mergeSocketDict[host]["command"]; ok {
			go commands.NewCommandProcess(host, dataSocket, commandSocket, &s.PassiveConnectManage)
			delete(s.mergeSocketDict, host)
		} else {
			s.mergeSocketDict[host] = map[string]net.Conn{
				"command": commandSocket,
			}
		}
	} else {

	}
}

// verifyDataSocket 判断对方是否已经通过预扫描验证
func (s *Server) verifyDataSocket(dataSocket net.Conn) {
	defer func(dataSocket net.Conn) {
		err := dataSocket.Close()
		if err != nil {
			logrus.Warning(err)
		}
	}(dataSocket)
	addr := dataSocket.RemoteAddr().String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	logrus.Infof("Starting to verify data socket connection from %s...", host)
	if hostInfo, ok := scan.VerifyManage[host]; ok && hostInfo.AesKey != "" {
		// todo: 验证通过处理
		if commandSocket, ok := s.mergeSocketDict[host]["command"]; ok {
			go commands.NewCommandProcess(host, dataSocket, commandSocket, &s.PassiveConnectManage)
			delete(s.mergeSocketDict, host)
		} else {
			s.mergeSocketDict[host] = map[string]net.Conn{
				"data": dataSocket,
			}
		}
	} else {

	}
}

// initClient 主动创建客户端连接对方
// 如果已经预验证，那么直接连接即可通过验证
func (s *Server) initClient(ip string) {
	c, ok := client.NewClient(ip, s.ActiveConnectManage)
	if ok {
		s.ActiveConnectManage[ip] = serverOption.ActiveConnectManage{
			ID:         c.RemoteID,
			CreateTime: time.Now(),
			Client:     c,
		}
	} else {
		c.Close()
	}
}
