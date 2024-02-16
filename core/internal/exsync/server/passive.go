package server

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/exsync/server/commands"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
)

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
	if hostInfo, ok := VerifyManage[host]; ok && hostInfo.AesKey != "" {
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
	if hostInfo, ok := VerifyManage[host]; ok && hostInfo.AesKey != "" {
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
