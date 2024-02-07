package base

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/socket"
	"EXSync/core/option"
	"EXSync/core/option/server/comm"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"os"
	"path"
	"sync"
)

// GetFile 根据相对路径列表在Space里搜索相应的文件，并获取到本机。
// relPaths的数量将决定以多少并发获取文件
// 如果本地文件是待续写文件，将会续写
// 如果本地文件不存在，将会创建
func (c *Base) GetFile(relPaths, outPaths []string, db *gorm.DB, space comm.UdDict) (ok bool, err error) {
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
		var file option.Index
		fileMark := hashext.GetRandomStr(8)
		replyMark := hashext.GetRandomStr(8)
		db.Where("path = ?", relPath).First(&file)
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
	session, err := socket.NewSession(c.TimeChannel, nil, c.CommandSocket, hashext.GetRandomStr(8), c.AesGCM)
	if err != nil {
		return false, err
	}
	defer session.Close()
	_, err = session.SendCommand(command, false, true)
	if err != nil {
		return false, err
	}

	// 初始化传输
	dataBlock := int64(config.PacketSize - 8 - c.EncryptionLoss)

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
		s, err := socket.NewSession(c.TimeChannel, c.DataSocket, nil, replyMark, c.AesGCM)
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
			logrus.Errorf("Client GetFile GetFile: File %s failed to retrieve from host %s", filePath, c.Ip)
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

				f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0667)
				if err != nil {
					return
				}
				defer f.Close()
				for {
					result, err := c.TimeChannel.Get(fileMark)
					if err != nil {
						logrus.Errorf("subRoutine: Timed out while reading data")
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
				f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0667)
				if err != nil {
					return
				}
				defer f.Close()
				for {
					result, err := c.TimeChannel.Get(fileMark)
					if err != nil {
						logrus.Errorf("subRoutine: Timed out while reading data")
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
	var wait sync.WaitGroup
	fileCount := len(relPaths)
	wait.Add(fileCount)
	for i := 0; i < fileCount; i++ {
		go subRoutine(i, &wait)
	}
	wait.Wait()

	return
}

func (c *Base) PostFile() {

}
