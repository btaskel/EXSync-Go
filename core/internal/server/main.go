package server

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/server/scan"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
)

//var socketManage map[string]any

type Server struct {
	scan.Scan
	StopNewConnections bool
}

func NewServer() *Server {
	server := Server{}
	go server.createDataSocket(config.Config.Server.Addr.Port)
	go server.createCommandSocket(config.Config.Server.Addr.Port + 1)

	return &server
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

func (s *Server) verifyDataSocket(dataSocket net.Conn) {
	logrus.Debugf("Starting to verify data socket connection from %s...", dataSocket.RemoteAddr().String())
}

func (s *Server) verifyCommandSocket(commandSocket net.Conn) {
	logrus.Debugf("Starting to verify command socket connection from %s...", commandSocket.RemoteAddr().String())
}
