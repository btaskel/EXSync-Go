package commands

import (
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	serverOption "EXSync/core/option/exsync/server"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type CommandProcess struct {
	AesGCM        *encryption.Gcm
	TimeChannel   *timechannel.TimeChannel
	CommandSocket net.Conn
	DataSocket    net.Conn
	IP            string

	close bool
}

// NewCommandProcess 对已经被动建立连接的CommandSocket进行初始化
func NewCommandProcess(host string, dataSocket, commandSocket net.Conn, passiveConnectManage map[string]serverOption.PassiveConnectManage, VerifyManage map[string]serverOption.VerifyManage) {
	// 判断是否有aes-gcm加密实例，如果有则在接收数据时使用aes-gcm；否则，明文传输；
	var gcm *encryption.Gcm
	hostVerifyInfo, ok := VerifyManage[host]
	if ok && len(hostVerifyInfo.AesKey) != 0 {
		var err error
		gcm, err = encryption.NewGCM(hostVerifyInfo.AesKey)
		if err != nil {
			logrus.Errorf("NewCommandProcess: Error creating instruction processor using %s! %s", hostVerifyInfo.AesKey, err)
			return
		}
	} else {
		gcm = nil
	}

	// 初始化TimeChannel与CommandProcess
	timeChannel := timechannel.NewTimeChannel()
	cp := CommandProcess{
		AesGCM:        gcm,
		TimeChannel:   timeChannel,
		DataSocket:    dataSocket,
		CommandSocket: commandSocket,
		IP:            host,
		close:         false,
	}

	// 将被动连接增加至PassiveConnectManage实例
	passiveConnectManage[host] = serverOption.PassiveConnectManage{
		ID:             "",
		CreateTime:     time.Now(),
		CommandProcess: &cp,
	}
	logrus.Debugf("NewCommandProcess: A passive connection for %s has been created.", host)

	var wait sync.WaitGroup
	wait.Add(2)
	go func() {
		cp.recvCommand() // 持续接收命令
		wait.Done()
	}()
	go func() {
		cp.recvData() // 持续接收数据(timeChannel)
		wait.Done()
	}()
	wait.Wait()

	// 如果在被动连接列表则进行关闭删除操作; 用于防止再次关闭
	if _, ok := passiveConnectManage[host]; ok {
		// 关闭socket
		cp.CommandSocket.Close()
		cp.DataSocket.Close()

		// 释放timeChannel
		cp.TimeChannel.Close()

		// 从被动连接列表删除已连接设备
		delete(passiveConnectManage, host)
	}

	logrus.Debugf("NewCommandProcess: Passive connection disconnected from host %s", host)
	return
}
