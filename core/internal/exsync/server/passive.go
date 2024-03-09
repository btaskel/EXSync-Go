package server

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/exsync/server/commands"
	loger "EXSync/core/log"
	serverOption "EXSync/core/option/exsync/manage"
	"fmt"
	"net"
)

// createCommandSocket 创建套接字对象
func (s *Server) createCommandSocket(port int) {
	address := fmt.Sprintf("%s:%d", config.Config.Server.Addr.IP, port)
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return
	}
	s.commListen = listen
	for {
		conn, err := listen.Accept()
		if err != nil {
			loger.Log.Debugf("address %s: %s", address, err)
			continue
		}
		go s.verifyCommandSocket(conn)
	}
}

// createDataSocket 创建套接字对象
func (s *Server) createDataSocket(port int) {
	address := fmt.Sprintf("%s:%d", config.Config.Server.Addr.IP, port)
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return
	}
	s.dataListen = listen
	for {
		conn, err := listen.Accept()
		if err != nil {
			loger.Log.Debugf("address %s: %s", address, err)
			continue
		}
		go s.verifyDataSocket(conn)
	}
}

// verifyCommandSocket 判断对方是否已经通过预扫描验证
func (s *Server) verifyCommandSocket(commandSocket net.Conn) {
	defer func(commandSocket net.Conn) {
		err := commandSocket.Close()
		if err != nil {
			loger.Log.Warning(err)
		}
	}(commandSocket)
	addr := commandSocket.RemoteAddr().String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	loger.Log.Infof("Starting to verify command socket connection from %s...", host)
	if hostInfo, ok := VerifyManage[host]; ok && hostInfo.AesKey != "" {
		if dataSocket, ok := s.mergeSocketDict[host]["command"]; ok {
			go commands.NewCommandProcess(host, dataSocket, commandSocket, s.PassiveConnectManage, VerifyManage, s.ctxServer, s.Trans)
			if _, ok := s.ActiveConnectManage[host]; !ok {
				go s.InitClient(host)
			}
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
			loger.Log.Warning(err)
		}
	}(dataSocket)
	addr := dataSocket.RemoteAddr().String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	loger.Log.Infof("Starting to verify data socket connection from %s...", host)
	if hostInfo, ok := VerifyManage[host]; ok && hostInfo.AesKey != "" {
		if commandSocket, ok := s.mergeSocketDict[host]["command"]; ok {
			go commands.NewCommandProcess(host, dataSocket, commandSocket, s.PassiveConnectManage, VerifyManage, s.ctxServer, s.Trans)
			if _, ok := s.ActiveConnectManage[host]; !ok {
				go s.InitClient(host)
			}
			delete(s.mergeSocketDict, host)
		} else {
			s.mergeSocketDict[host] = map[string]net.Conn{
				"data": dataSocket,
			}
		}
	} else {

	}
}

// ClosePassiveConnect 关闭一个被动连接
func (s *Server) ClosePassiveConnect(cp *commands.CommandProcess, passiveConnectManage map[string]serverOption.PassiveConnectManage) bool {
	addr := cp.CommandSocket.RemoteAddr().String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	// 关闭socket
	err = cp.CommandSocket.Close()
	if err != nil {
		loger.Log.Warningf("Passive-ClosePassiveConnect: An error occurred when disconnecting from the CommandSocket connection of the %s host", host)
		return false
	}
	err = cp.DataSocket.Close()
	if err != nil {
		loger.Log.Warningf("Passive-ClosePassiveConnect: An error occurred when disconnecting from the DataSocket connection of the %s host", host)
		return false
	}

	// 释放timeChannel
	cp.TimeChannel.Close()

	// 从被动连接列表删除已连接设备
	delete(passiveConnectManage, host)
	return true
}
