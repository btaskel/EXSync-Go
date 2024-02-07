package base

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/socket"
	"EXSync/core/option"
	"EXSync/core/option/server/comm"
	"encoding/json"
	"fmt"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io"
	"os"
	"path"
	"sync"
)

func (c *Base) sendData(file *os.File, fileMark string, dataBlock, total int64) {
	// total = 将会读取多少数据 = remoteSize - localSize
	s, err := socket.NewSession(nil, c.DataSocket, nil, fileMark, c.AesGCM)
	if err != nil {
		return
	}

	buffer := make([]byte, dataBlock)
	var readData int64
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
			return
		}
		readData += dataBlock
	}
}

func (c *Base) GetFile(data map[string]any) {
	pathList, ok := data["pathList"].([]string)
	if !ok {
		return
	}
	hashList, ok := data["hashList"].([]string)
	if !ok {
		return
	}
	sizeList, ok := data["sizeList"].([]int64)
	if !ok {
		return
	}
	markList, ok := data["markList"].([]string)
	if !ok {
		return
	}
	replyMarkList, ok := data["replyMarkList"].([]string)
	if !ok {
		return
	}
	spaceName, ok := data["spaceName"].(string)
	if !ok {
		return
	}

	// 连接数据库
	spacePath := config.UserData[spaceName].Path
	dbPath := path.Join(spacePath, config.SpaceInfoPath, "files.db")
	sqliteOpen := sqlite.Open(dbPath)
	db, err := gorm.Open(sqliteOpen, &gorm.Config{})
	if err != nil {
		return
	}

	dataBlock := int64(config.PacketSize - 8 - c.EncryptionLoss)

	subRoutine := func(i int, wait *sync.WaitGroup) {
		defer wait.Done()
		remoteRelPath := pathList[i]
		remoteAbsPath := path.Join(spacePath, remoteRelPath)
		remoteFileHash := hashList[i]
		remoteFileSize := sizeList[i]
		fileMark := markList[i]
		replyMark := replyMarkList[i]

		session, err := socket.NewSession(c.TimeChannel, c.DataSocket, nil, replyMark, c.AesGCM)
		if err != nil {
			return
		}
		defer session.Close()

		var file option.Index
		db.Where("path = ?", remoteRelPath).First(&file)

		command := comm.Command{
			Command: "",
			Type:    "",
			Method:  "",
			Data: map[string]any{
				"size": file.Size,
				"hash": file.Hash,
			},
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

				var continueWrite bool
				if hashList[i] == fmt.Sprintf("%X", hasher.Sum(nil)) {
					// 判断为需要续写的文件
					continueWrite = true
				} else {
					continueWrite = true
				}
				command = comm.Command{
					Command: "",
					Type:    "",
					Method:  "",
					Data: map[string]any{
						"ok": continueWrite,
					},
				}
				_, err = session.SendCommand(command, false, true)
				if continueWrite {
					f, err = os.Open(remoteAbsPath)
					if err != nil {
						return
					}
					defer f.Close()
					c.sendData(f, fileMark, dataBlock, file.Size-remoteFileSize)
				}

			} else if file.Size > remoteFileSize && remoteFileSize == 0 {
				// 远程文件不存在，需要传输
				f, err := os.Open(remoteAbsPath)
				if err != nil {
					return
				}
				defer f.Close()
				c.sendData(f, fileMark, dataBlock, file.Size-remoteFileSize)
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

	var wait sync.WaitGroup
	fileCount := len(pathList)
	wait.Add(fileCount)
	for i := 0; i > len(pathList); i++ {
		go subRoutine(i, &wait)
	}
	wait.Wait()
}

//// GetFile 客户端从服务端接收数据
//// 发送文件至客户端
//// mode = 0;
//// 直接发送所有数据。
////
//// mode = 1;
//// 根据客户端发送的文件哈希值，判断是否是意外中断传输的文件，如果是则继续传输。
//func (c *Base) GetFile(data map[string]any, replyMark string, db *gorm.DB) {
//	fileRelPath, ok := data["path"].(string)
//	if !ok {
//		return
//	}
//	remoteFileHash, ok := data["local_file_hash"].(string)
//	if !ok {
//		return
//	}
//	remoteSize, ok := data["local_file_size"].(int64)
//	if !ok {
//		return
//	}
//	fileMark, ok := data["fileMark"].(string)
//	if !ok {
//		return
//	}
//	spacename, ok := data["spacename"].(string)
//	if !ok {
//		return
//	}
//
//	localSpace := config.UserData[spacename]
//	fileAbsPath := path.Join(localSpace.Path, fileRelPath)
//
//	dataBlock := int64(config.PacketSize - 8 - c.EncryptionLoss) // fileMark 8 + GCM-tag 16 + GCM-nonce 12
//
//	status := "ok"
//
//	var file option.Index
//	var command comm.Command
//	db.Where("path = ?", fileRelPath).First(&file)
//	if file.Path == "" {
//		command = comm.Command{
//			Command: "",
//			Type:    "",
//			Method:  "",
//			Data: map[string]any{
//				"local_file_size": 0,
//				"local_file_hash": 0,
//			},
//		}
//	} else {
//		command = comm.Command{
//			Command: "",
//			Type:    "",
//			Method:  "",
//			Data: map[string]any{
//				"local_file_size": file.Size,
//				"local_file_hash": file.Hash,
//			},
//		}
//	}
//	session, err := socket.NewSession(c.TimeChannel, c.DataSocket, nil, replyMark, c.AesGCM)
//	if err != nil {
//		return
//	}
//	_, err = session.SendCommand(command, false, false)
//	if err != nil {
//		return
//	}
//
//	// 检查是否需要续写文件
//	// 远程文件大小小于本地文件
//
//	sendData := func(f *os.File) {
//		s, err := socket.NewSession(nil, c.DataSocket, nil, fileMark, c.AesGCM)
//		if err != nil {
//			return
//		}
//		buffer := make([]byte, dataBlock)
//		for {
//			n, err := f.Read(buffer)
//			if err != nil && err != io.EOF {
//				return
//			}
//			if n == 0 {
//				break
//			}
//			err = s.SendDataP(buffer)
//			if err != nil {
//				return
//			}
//		}
//		f.Close()
//	}
//
//	if remoteSize < file.Size {
//		fileBlock, littleBlock := remoteSize/8192, remoteSize%8192
//		f, err := os.Open(fileAbsPath)
//		if err != nil {
//			return
//		}
//		hasher, err := hashext.UpdateXXHash(f, int(fileBlock))
//		if err != nil {
//			return
//		}
//		buf := make([]byte, littleBlock)
//		n, err := f.Read(buf)
//		_, err = hasher.Write(buf[:n])
//		if err != nil {
//			return
//		}
//		f.Close()
//		if remoteFileHash == fmt.Sprintf("%X", hasher.Sum(nil)) {
//			command = comm.Command{
//				Command: "",
//				Type:    "",
//				Method:  "",
//				Data: map[string]any{
//					"ok":     true,
//					"status": status,
//				},
//			}
//			_, err = session.SendCommand(command, false, true)
//			if err != nil {
//				return
//			}
//
//			// 准备发送待续传的数据
//			f, err = os.Open(fileAbsPath)
//			if err != nil {
//				return
//			}
//
//			//0（os.SEEK_SET）：表示相对于文件开始的位置。
//			//1（os.SEEK_CUR）：表示相对于当前位置。
//			//2（os.SEEK_END）：表示相对于文件结束的位置。
//			_, err = f.Seek(remoteSize, 1)
//			if err != nil {
//				return
//			}
//			sendData(f)
//		}
//	} else if remoteSize > file.Size {
//		f, err := os.Open(fileAbsPath)
//		if err != nil {
//			return
//		}
//		sendData(f)
//	} else {
//		if remoteFileHash == file.Hash {
//			f, err := os.Open(fileAbsPath)
//			if err != nil {
//				return
//			}
//			sendData(f)
//		} else {
//			// 文件大小不相同
//			return
//		}
//	}
//}

// PostFile 客户端发送文件至服务端
func (c *Base) PostFile(data map[string]any, replyMark string, db *gorm.DB) {
	remoteSpaceName, ok := data["spacename"].(string)
	if !ok {
		return
	}
	localSpace, ok := config.UserData[remoteSpaceName]
	if !ok {
		return
	}
	remoteFileRelPath, ok := data["file_path"].(string)
	if !ok {
		return
	}
	remoteFileSize, ok := data["file_size"].(int64)
	if !ok {
		return
	}
	remoteFileHash, ok := data["file_hash"].(string)
	if !ok {
		return
	}
	mode, ok := data["mode"].(int)
	if !ok {
		return
	}
	fileMark, ok := data["filemark"].(string)
	if !ok {
		return
	}

	// 初始化接收数据队列
	err := c.TimeChannel.CreateRecv(fileMark)
	if err != nil {
		return
	}

	var status = ""

	fileAbsPath := path.Join(localSpace.Path, remoteFileRelPath)
	if len(remoteFileHash) != 32 {
		status = "File hash too long!"
	}

	// 获取db
	//spaceDb, err := index.Index.GetIndex(path.Join(localSpace.Path, ".\\sync\\info\\files.db"))
	//if err != nil {
	//	logrus.Errorf("Host %s failed to open the index database for %s!", c.Ip, remoteSpace)
	//}

	var file option.Index
	var command comm.Command
	db.Where("Path = ?", remoteFileRelPath).First(&file)
	//spaceIndex.First(&indexOption, "code = ?", "D42")

	//if result.Error != nil {
	//	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
	//		size = 0
	//	} else {
	//		logrus.Errorf("")
	//	}
	//}
	if file.Path == "" {
		command = comm.Command{
			Command: "",
			Type:    "",
			Method:  "",
			Data: map[string]any{
				"file_size": 0,
				"file_hash": "",
				"file_date": 0,
				"status":    status,
			},
		}
	} else {
		command = comm.Command{
			Command: "",
			Type:    "",
			Method:  "",
			Data: map[string]any{
				"file_size": file.Size,
				"file_hash": file.Hash,
				"file_date": file.EditDate,
				"status":    status,
			},
		}
	}

	// 创建会话
	session, err := socket.NewSession(c.TimeChannel, c.DataSocket, nil, replyMark, c.AesGCM)
	if err != nil {
		return
	}

	_, err = session.SendCommand(command, false, true)
	if err != nil {
		return
	}

	dataBlock := int64(config.PacketSize - 8 - c.EncryptionLoss) // fileMark 8 + GCM-tag 16 + GCM-nonce 12
	switch mode {
	case 0:
		// 如果不存在文件，则创建文件。否则不执行操作。
		if file.Path == "" {
			readData := remoteFileSize
			f, err := os.OpenFile(fileAbsPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				logrus.Errorf("PostFile: Error opening or creating file:%s", err)
				return
			}

			for readData > 0 {
				readData -= dataBlock
				result, err := c.TimeChannel.Get(fileMark)
				if err != nil {
					f.Close()
					return
				}
				_, err = f.Write(result)
				if err != nil {
					f.Close()
					return
				}
			}
			f.Close()
			return
		} else {
			return
		}
	case 1:
		// 如果不存在文件，则创建文件。否则重写文件。
		if file.Path != "" {
			err = os.Remove(fileAbsPath)
			if err != nil {
				// 数据不同步导致问题
				logrus.Debugf("Error removing file:%s", err)
				return
			}
		}

		f, err := os.OpenFile(fileAbsPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return
		}
		readData := remoteFileSize
		for readData > 0 {
			readData -= dataBlock
			result, err := c.TimeChannel.Get(fileMark)
			if err != nil {
				f.Close()
				return
			}
			_, err = f.Write(result)
			if err != nil {
				f.Close()
				return
			}
		}
	case 2:
		// 如果存在文件，并且准备发送的文件字节是对方文件字节的超集(xxh3_128相同)，则续写文件。
		if file.Path == "" {
			return
		}
		// 是否要进行续传
		reply, err := c.TimeChannel.GetTimeout(replyMark, int(remoteFileSize/1048576))
		if err != nil {
			return
		}

		var replyCommand comm.Command
		err = json.Unmarshal(reply, &replyCommand)
		if err != nil {
			return
		}

		fileStatus := replyCommand.Data["status"].(bool)
		if fileStatus {
			f, err := os.OpenFile(fileAbsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return
			}

			difference := remoteFileSize - file.Size
			var readData int64
			for readData <= difference {
				result, err := c.TimeChannel.Get(fileMark)
				if err != nil {
					return
				}
				_, err = f.Write(result)
				if err != nil {
					return
				}
				readData += dataBlock
			}
			return
		}
	}
}
