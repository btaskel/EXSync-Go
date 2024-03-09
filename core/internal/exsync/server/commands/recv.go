package commands

import (
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	loger "EXSync/core/log"
	"EXSync/core/option/exsync/manage"
	serverOption "EXSync/core/option/exsync/server"
	"EXSync/core/option/exsync/trans"
	"context"
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
	ctxProcess    context.Context // 上下文管理
	taskManage    map[string]trans.TranTask
}

// NewCommandProcess 对已经被动建立连接的CommandSocket进行初始化
func NewCommandProcess(host string, dataSocket, commandSocket net.Conn, passiveConnectManage map[string]manage.PassiveConnectManage,
	VerifyManage map[string]serverOption.VerifyManage, ctxServer context.Context, taskManage map[string]trans.TranTask) {
	// 判断是否有aes-gcm加密实例，如果有则在接收数据时使用aes-gcm；否则，明文传输；
	var gcm *encryption.Gcm
	hostVerifyInfo, ok := VerifyManage[host]
	if ok && len(hostVerifyInfo.AesKey) != 0 {
		var err error
		gcm, err = encryption.NewGCM(hostVerifyInfo.AesKey)
		if err != nil {
			loger.Log.Errorf("NewCommandProcess: Error creating instruction processor using %s! %s", hostVerifyInfo.AesKey, err)
			return
		}
	} else {
		gcm = nil
	}

	// 创建上下文分支
	ctxProcess, cancelProcess := context.WithCancel(ctxServer)

	// 初始化TimeChannel与CommandProcess
	timeChannel := timechannel.NewTimeChannel()
	cp := CommandProcess{
		AesGCM:        gcm,
		TimeChannel:   timeChannel,
		DataSocket:    dataSocket,
		CommandSocket: commandSocket,
		IP:            host,
		ctxProcess:    ctxProcess,
		taskManage:    taskManage,
	}

	// 将被动连接增加至PassiveConnectManage实例
	passiveConnectManage[host] = manage.PassiveConnectManage{
		ID:             "",
		CreateTime:     time.Now(),
		CommandProcess: &cp,
		Cancel:         cancelProcess,
	}
	loger.Log.Debugf("NewCommandProcess: A passive connection for %s has been created.", host)

	var wg sync.WaitGroup
	wg.Add(2)
	go cp.recvCommand(&wg) // 持续接收命令
	go cp.recvData(&wg)    // 持续接收数据(timeChannel)
	wg.Wait()

	// 如果在被动连接列表则进行关闭删除操作; 用于防止再次关闭
	if _, ok = passiveConnectManage[host]; ok {
		// 关闭socket
		cp.CommandSocket.Close()
		cp.DataSocket.Close()

		// 释放timeChannel
		cp.TimeChannel.Close()

		// 从被动连接列表删除已连接设备
		delete(passiveConnectManage, host)
	}

	loger.Log.Debugf("NewCommandProcess: Passive connection disconnected from host %s", host)
	return
}
