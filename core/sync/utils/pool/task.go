package pool

import (
	"EXSync/core/internal/exsync/client"
	"EXSync/core/internal/exsync/server"
	loger "EXSync/core/log"
	clientComm "EXSync/core/option/exsync/comm/client"
)

func (p *Pool) executeFunc(c *client.Client, files map[string]file, hostStress HostStress, hostAddr string) {
	defer func() {
		// 释放该批次产生的主机压力值
		for _, f := range files {
			hostStress.total -= f.TotalSize
		}
	}()

	// 构建Client GetFile所需的参数格式
	var fs []clientComm.GetFile
	for relPath, info := range files {
		f := clientComm.GetFile{
			RelPath: relPath,
			OutPath: info.OutPath,
			Size:    info.FileSize,
			Date:    info.FileDate,
			Hash:    info.FileHash,
			Space:   info.Space,
			Offset:  info.Offset,
		}
		fs = append(fs, f)
	}

	errResultChan := make(chan map[string]error)
	errChan := make(chan error)

	go func() {
		failedFiles, err := c.Comm.GetFile(p.ctx, fs)
		if err != nil {
			errChan <- err
		} else {
			errResultChan <- failedFiles
		}
	}()

	select {
	case <-p.ctx.Done():
		loger.Log.Infof("executeFunc -> %s: Canceled GetFile operation", c.IP)
		return
	case failedFiles := <-errResultChan:
		for failedFileName, failedErr := range failedFiles {
			loger.Log.Warningf("executeFunc -> %s: File %s Get failed, %s!", c.IP, failedFileName, failedErr)

			hosts := files[failedFileName].Hosts
			if len(hosts) == 1 {
				// todo: 没有备用主机, 取消传输
				server.Fails <- map[string]error{failedFileName: failedErr}
			}
			for n, host := range hosts {
				if host == hostAddr {
					fileTask := files[failedFileName]
					fileTask.Hosts = append(fileTask.Hosts[:n], fileTask.Hosts[n+1:]...)

					// 重新追加到waitQueue
					p.waitQueue <- fileTask
					break
				}
			}
		}
		return
	case errInfo := <-errChan:
		loger.Log.Warningf("executeFunc -> %s: %s !", c.IP, errInfo)
		return
	}

}
