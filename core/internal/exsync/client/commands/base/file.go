package base

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/pathext"
	"EXSync/core/internal/modules/socket"
	"EXSync/core/internal/modules/sqlt"
	configOption "EXSync/core/option/config"
	"EXSync/core/option/exsync/comm"
	"errors"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"sync"
)

// GetFile 根据相对路径列表在Space里搜索相应的文件，并获取到本机。
// relPaths的数量将决定以多少并发获取文件
// 如果本地文件是待续写文件，将会续写
// 如果本地文件不存在，将会创建
func (b *Base) GetFile(relPaths, outPaths []string, space configOption.UdDict) (failedFiles []string, err error) {
	// 初始化GetFile
	var (
		hashList      []string
		sizeList      []int64
		replyMarkList []string
		fileMarkList  []string
	)

	if outPaths == nil {
		outPaths = relPaths
	} else if outPaths != nil && len(outPaths) != len(outPaths) {
		panic("GetFile:获取的文件路径与保存路径数量不同步！")
	}

	for _, relPath := range relPaths {
		fileMark := hashext.GetRandomStr(8)
		replyMark := hashext.GetRandomStr(8)
		//db.Where("path = ?", relPath).First(&file)
		file, err := sqlt.QueryFile(space.Db, relPath)
		if err != nil {
			logrus.Errorf("Active GetFile: Error querying information for file %s!", relPath)
			failedFiles = append(failedFiles, relPath)
		}
		fileStat, err := os.Stat(path.Join(space.Path, relPath))
		if err == nil {
			fs := fileStat.Size()
			if fs != 0 && fs != file.Size {
				// 索引与本地文件实际情况不同步
				return
			}
		}

		hashList = append(hashList, file.Hash)
		sizeList = append(sizeList, file.Size)
		replyMarkList = append(replyMarkList, replyMark)
		fileMarkList = append(fileMarkList, fileMark)
	}

	// 准备数据
	command := comm.Command{
		Command: "data",
		Type:    "file",
		Method:  "get",
		Data: map[string]any{
			"pathList":      relPaths,
			"hashList":      hashList,
			"sizeList":      sizeList,
			"markList":      fileMarkList,
			"replyMarkList": replyMarkList,
			"spaceName":     space.SpaceName,
		},
	}
	session, err := socket.NewSession(b.TimeChannel, nil, b.CommandSocket, hashext.GetRandomStr(8), b.AesGCM)
	if err != nil {
		return relPaths, err
	}
	defer session.Close()
	_, err = session.SendCommand(command, false, true)
	if err != nil {
		return relPaths, err
	}

	// 初始化传输
	dataBlock := int64(config.PacketSize - 8 - b.EncryptionLoss)

	// 单文件传输线程
	subRoutine := func(i int, wait *sync.WaitGroup) {
		defer wait.Done()

		localFileHash := hashList[i]
		localFileSize := sizeList[i]
		fileMark := fileMarkList[i]
		replyMark := replyMarkList[i]

		// 设置输出路径
		var outPath string
		if len(relPaths) != 1 {
			outPath = relPaths[i]
		} else {
			outPath = relPaths[0]
		}

		filePath := path.Join(space.Path, outPath)

		// 创建会话接收首次答复
		s, err := socket.NewSession(b.TimeChannel, b.DataSocket, nil, replyMark, b.AesGCM)
		defer s.Close()
		if err != nil {
			return
		}
		reply, ok := s.Recv()
		if !ok {
			return
		}

		// 处理远程文件状态
		remoteFileSize, ok := reply.Data["size"].(int64)
		if !ok {
			return
		}
		remoteFileHash, ok := reply.Data["hash"].(string)

		if remoteFileSize == 0 {
			logrus.Errorf("active Client GetFile: File %s failed to retrieve from host %s", filePath, b.Ip)
			return
		}

		if remoteFileHash != localFileHash {
			if remoteFileSize > localFileSize && localFileSize != 0 {
				// 1.远程文件大于本地文件，等待对方判断是否为需续写文件；
				// 2.本地文件不存在，需要对方发送文件；

				reply, ok = s.RecvTimeout(int(remoteFileSize / 1048576))
				if !ok {
					return
				}
				pass, ok := reply.Data["ok"].(bool)
				if !ok || !pass {
					return
				}

				pathext.MakeDir(filePath)
				f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0667)
				if err != nil {
					return
				}
				defer f.Close()
				for {
					result, err := b.TimeChannel.GetTimeout(fileMark, config.SocketTimeout)
					if err != nil {
						if err == errors.New("timeout") {
							logrus.Errorf("active Sync Space %s :Receiving %s file timeout!", space.SpaceName, filePath)
						} else {
							logrus.Errorf("active Sync Space %s :Unknown error receiving %s file from host %s", space.SpaceName, filePath, b.Ip)
						}
						return
					}
					_, err = f.Write(result)
					if err != nil {
						return
					}
					localFileSize += dataBlock
					if localFileSize >= remoteFileSize {
						break
					}
				}
				return

			} else if remoteFileSize > localFileSize && localFileSize == 0 {
				// 本地文件不存在
				pathext.MakeDir(filePath)
				f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0667)
				if err != nil {
					return
				}
				defer f.Close()
				// 启用缓存
				for {
					result, err := b.TimeChannel.GetTimeout(fileMark, config.SocketTimeout)
					if err != nil {
						if err == errors.New("timeout") {
							logrus.Errorf("active Sync Space %s :Receiving %s file timeout!", space.SpaceName, filePath)
						} else {
							logrus.Errorf("active Sync Space %s :Unknown error receiving %s file from host %s", space.SpaceName, filePath, b.Ip)
						}
						return
					}
					_, err = f.Write(result)
					if err != nil {
						return
					}
					localFileSize += dataBlock
					if localFileSize >= remoteFileSize {
						break
					}
				}

			} else if remoteFileSize == localFileSize {
				// 1.文件大小相同，哈希值不同；
				// 2.本地或远程文件不存在；
				if remoteFileSize == 0 {
					// 远程文件不存在
					return
				} else {
					// 文件不同
					return
				}

			} else if remoteFileSize < localFileSize {
				// 1.远程文件小于本地文件，不进行同步
				// 2.在获取一个远程不存在的文件
			}
		} else {
			// 正在尝试获取一个同样的文件，没有意义。
		}
	}

	// 开始传输
	logrus.Debugf("active Sync Space %s :File %s begins to transfer", space.SpaceName, relPaths)
	var wait sync.WaitGroup
	fileCount := len(relPaths)
	wait.Add(fileCount)
	for i := 0; i < fileCount; i++ {
		go subRoutine(i, &wait)
	}
	wait.Wait()
	logrus.Debugf("active Sync Space %s :File %s transfer completed", space.SpaceName, relPaths)
	return
}

func (b *Base) PostFile() {

}
