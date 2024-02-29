package pool

import (
	"EXSync/core/internal/exsync/client"
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

	failedFiles, err := c.Comm.GetFile(fs)
	if err != nil {
		// 批量拉取文件发送错误
		loger.Log.Warningf("pool: %s", err)
		return
	}

	for failedFileName, failedErr := range failedFiles {
		loger.Log.Warningf("")
		// todo: 此处汇报上层

		hosts := files[failedFileName].Hosts
		if len(hosts) == 1 {
			// todo: 没有备用主机, 取消传输
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
}
