package base

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/socket"
	"EXSync/core/internal/modules/sqlt"
	loger "EXSync/core/log"
	"EXSync/core/option/exsync/comm"
	"EXSync/core/option/exsync/trans"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
)

func (c *Base) GetFile(data map[string]any) {
	// 权限验证
	if !CheckPermission(c.VerifyManage[c.Ip], []string{comm.PermRead}) {
		loger.Log.Warningf("Passive-GetFile-Permission: Host %s is attempting an unauthorized operation <read> !", c.Ip)
		return
	}

	files, ok := data["files"].(map[string]any)
	if !ok {
		return
	}

	subRoutine := func(remoteRelPath string, fileInfo map[string]any, wait *sync.WaitGroup) {
		defer wait.Done()
		var progress float32
		c.TaskManage[remoteRelPath] = trans.TranTask{
			DstAddr:  c.Ip,
			Progress: &progress,
			Cancel:   nil,
		}

		remoteOffset, ok := fileInfo["offset"].(int64)
		if ok {
			//Ctx := c.CtxProcess
			c.direct(fileInfo, remoteRelPath, remoteOffset, &progress)
		} else {
			c.check(fileInfo, remoteRelPath, &progress)
		}
	}

	var wait sync.WaitGroup
	fileCount := len(files)
	wait.Add(fileCount)
	for filePath, fileInfo := range files {
		go subRoutine(filePath, fileInfo.(map[string]any), &wait)
	}
	wait.Wait()
	loger.Log.Debugf("passive-GetFile :File a, b begins to transfer")
}

// check 检查并判断文件状态进行传输
func (c *Base) check(fileInfo map[string]any, remoteRelPath string, progress *float32) {
	sendData := func(file *os.File, fileMark string, dataBlock, total int64) (err error) {
		// total = 将会读取多少数据 = remoteSize - localSize
		s, err := socket.NewSession(nil, c.DataSocket, nil, fileMark, c.AesGCM)
		defer s.Close()
		if err != nil {
			return
		}

		var readData int64
		buffer := make([]byte, dataBlock)
		for {
			n, err := file.Read(buffer)
			if err != nil && err != io.EOF {
				return
			}
			if n == 0 || readData >= total {
				break
			}
			err = s.SendDataP(buffer)
			if err != nil {
				return err
			}
			v := (float32(readData) / float32(total)) * 100
			progress = &v
			readData += dataBlock
		}
		return
	}

	dataBlock := int64(config.PacketSize - 8 - c.EncryptionLoss)

	remoteFileHash, ok := fileInfo["hash"].(string)
	if !ok {
		return
	}
	remoteFileSize, ok := fileInfo["size"].(int64)
	if !ok {
		return
	}
	fileMark, ok := fileInfo["fileMark"].(string)
	if !ok {
		return
	}
	replyMark, ok := fileInfo["replyMark"].(string)
	if !ok {
		return
	}
	spaceName, ok := fileInfo["spaceName"].(string)
	if !ok {
		return
	}

	session, err := socket.NewSession(c.TimeChannel, c.DataSocket, nil, replyMark, c.AesGCM)
	if err != nil {
		return
	}
	defer session.Close()

	// 连接数据库
	space, ok := config.UserData[spaceName]
	if !ok {
		return
	}
	var stat string
	file, err := sqlt.QueryFile(space.Db, remoteRelPath)
	if err != nil {
		stat = "Passive-GetFile: QueryFile Error!"
	}

	remoteAbsPath := filepath.Join(space.Path, remoteRelPath)
	// 发送异常或者文件信息
	var command comm.Command
	if len(stat) != 0 {
		command = comm.Command{
			Data: map[string]any{
				"size": file.Size,
				"hash": file.Hash,
				"date": file.EditDate,
			},
		}
	} else {
		command = comm.Command{
			Data: map[string]any{
				"stat": stat,
			},
		}
	}
	_, err = session.SendCommand(command, false, true)
	if err != nil {
		return
	}
	if remoteFileHash != file.Hash {
		if file.Size > remoteFileSize && remoteFileSize != 0 {
			// 1.本地文件大于远程文件，等待判断是否为需续写文件；
			// 2.本地文件不存在，需要对方发送文件；
			f, err := os.Open(remoteAbsPath)
			if err != nil {
				return
			}
			fileBlock, littleBlock := file.Size/8192, file.Size%8192
			hasher, err := hashext.UpdateXXHash(f, int(fileBlock))
			if err != nil {
				return
			}
			buf := make([]byte, littleBlock)
			n, err := f.Read(buf)
			if err != nil {
				return
			}
			_, err = hasher.Write(buf[:n])
			if err != nil {
				return
			}

			// 判断为需要续写的文件
			var continueWrite bool
			if remoteFileHash == fmt.Sprintf("%X", hasher.Sum(nil)) {
				continueWrite = true
			} else {
				continueWrite = false
			}

			reply := comm.Command{
				Data: map[string]any{
					"ok": continueWrite,
				},
			}
			_, err = session.SendCommand(reply, false, true)
			if err != nil {
				return
			}

			if continueWrite {
				f, err = os.Open(remoteAbsPath)
				if err != nil {
					return
				}
				defer func(f *os.File) {
					err = f.Close()
					if err != nil {
						return
					}
				}(f)
				err = sendData(f, fileMark, dataBlock, file.Size-remoteFileSize)
				if err != nil {
					netErr, ok := err.(net.Error)
					if ok && netErr.Timeout() {
						loger.Log.Errorf("passive-GetFile - Sync Space %s :Sending %s file timeout!", space.SpaceName, remoteAbsPath)
						return
					} else {
						loger.Log.Warningf("passive-GetFile - Sync Space %s :File %s transfer failed due to unexpected disconnection from host %s.", space.SpaceName, remoteAbsPath, c.Ip)
						return
					}
				}
			}

		} else if file.Size > remoteFileSize && remoteFileSize == 0 {
			// 远程文件不存在，需要传输
			f, err := os.Open(remoteAbsPath)
			if err != nil {
				return
			}
			defer func(f *os.File) {
				err = f.Close()
				if err != nil {
					return
				}
			}(f)
			err = sendData(f, fileMark, dataBlock, file.Size-remoteFileSize)
			if err != nil {
				netErr, ok := err.(net.Error)
				if ok && netErr.Timeout() {
					loger.Log.Errorf("passive-GetFile - Sync Space %s :Sending %s file timeout!", space.SpaceName, remoteAbsPath)
					return
				} else {
					loger.Log.Warningf("passive-GetFile - Sync Space %s :File %s transfer failed due to unexpected disconnection from host %s.", space.SpaceName, remoteAbsPath, c.Ip)
					return
				}
			}
		} else if file.Size == remoteFileSize {
			// 1.文件大小相同，哈希值不同；
			// 2.本地或远程文件不存在；
		} else {
			// 1.远程文件小于本地文件，不进行同步
			// 2.在获取一个远程不存在的文件
		}
	} else {
		// 正在尝试获取一个同样的文件，没有意义。
	}
}

// direct 根据偏移量进行传输
func (c *Base) direct(fileInfo map[string]any, remoteRelPath string, remoteOffset int64, progress *float32) {
	dataBlock := int64(config.PacketSize - 8 - c.EncryptionLoss)

	fileMark, ok := fileInfo["fileMark"].(string)
	if !ok {
		loger.Log.Errorf("Passive-GetFile-Direct: Missing parameter <fileMark>!")
		return
	}
	spaceName, ok := fileInfo["spaceName"].(string)
	if !ok {
		loger.Log.Errorf("Passive-GetFile-Direct: Missing parameter <spaceName>!")
		return
	}

	// 查询本地同步空间
	space, ok := config.UserData[spaceName]
	if !ok {
		loger.Log.Errorf("Passive-GetFile-Direct: syncSpace %s not exist!", spaceName)
		return
	}

	s, err := socket.NewSession(c.TimeChannel, c.DataSocket, nil, fileMark, c.AesGCM)
	defer s.Close()
	if err != nil {
		loger.Log.Errorf("Passive-GetFile-Direct: Create session failed! %s", err)
		return
	}

	remoteAbsPath := filepath.Join(space.Path, remoteRelPath)

	f, err := os.Open(remoteAbsPath)
	if err != nil {
		loger.Log.Errorf("Passive-GetFile-Direct: Local file %s read failure!", remoteAbsPath)
		return
	}
	_, err = f.Seek(remoteOffset, 0)
	if err != nil {
		loger.Log.Errorf("Passive-GetFile-Direct: Local file %s offset failed!", remoteAbsPath)
		return
	}
	stat, err := f.Stat()
	if err != nil {
		return
	}
	fileSize := stat.Size()

	var readData int64
	buf := make([]byte, dataBlock)
	for {
		n, err := f.Read(buf)
		readData += int64(n)
		if err != nil && err == io.EOF {
			loger.Log.Debugf("Passive-GetFile-Direct: Local file %s successfully transferred!", remoteAbsPath)
			return
		}

		err = s.SendDataP(buf[:n])
		if err != nil {
			return
		}
		v := (float32(readData) / float32(fileSize)) * 100
		progress = &v
	}
}
