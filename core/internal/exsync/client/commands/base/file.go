package base

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/buffer"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/pathext"
	"EXSync/core/internal/modules/socket"
	loger "EXSync/core/log"
	"EXSync/core/option/exsync/comm"
	clientComm "EXSync/core/option/exsync/comm/client"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
)

var (
	FileQueryErr  = errors.New("FileQueryErr")  // 查询数据不存在
	FileSyncDBErr = errors.New("FileSyncDBErr") // 文件与数据库不同步
	FileOpenErr   = errors.New("FileOpenErr")   // 文件打开失败
	FileCloseErr  = errors.New("FileCloseErr")  // 文件关闭失败

	SocketNewSessionErr = errors.New("SocketNewSessionErr") // 创建Session对象失败
	SocketTimeoutErr    = errors.New("SocketTimeoutErr")    // 接收数据超时
	SocketBufferErr     = errors.New("SocketBufferErr")     // 文件写入缓冲遇到错误

	ParamsNotExistsErr = errors.New("ParamsNotExistsErr") // 缺少必要的参数
)

// GetFile 根据相对路径列表在Space里搜索相应的文件，并获取到本机。
// relPaths的数量将决定以多少并发获取文件
// offset = 0 :本地文件是待续写文件，将会续写；如果本地文件不存在，将会创建
// offset ≠ 0 :直接根据偏移量获取指定的大小
func (b *Base) GetFile(inputFiles []clientComm.GetFile) (failedFiles map[string]error, err error) {
	// 权限验证
	if !CheckPermission(b.VerifyManage, []string{comm.PermRead}) {
		return nil, errors.New("MissingPermissions")
	}

	// 准备文件信息
	var files map[string]any
	for _, inputFile := range inputFiles {
		if fileStat, err := os.Stat(filepath.Join(inputFile.Space.Path, inputFile.RelPath)); err == nil {
			fs := fileStat.Size()
			if fs != inputFile.Size {
				// 索引与本地文件实际情况不同步
				failedFiles[inputFile.RelPath] = FileSyncDBErr
			}
		}

		files[inputFile.RelPath] = map[string]any{
			"hash":      inputFile.Hash,
			"size":      inputFile.Size,
			"replyMark": hashext.GetRandomStr(6),
			"fileMark":  hashext.GetRandomStr(6),
			"spaceName": inputFile.Space.SpaceName,
			"offset":    inputFile.Offset,
		}
	}

	// 如果准备文件信息有错误, 则返回错误map
	if len(failedFiles) != 0 {
		return
	}

	// 发送数据
	command := comm.Command{
		Command: "data",
		Type:    "file",
		Method:  "get",
		Data: map[string]any{
			"files": files,
		},
	}
	session, err := socket.NewSession(b.TimeChannel, nil, b.CommandSocket, hashext.GetRandomStr(6), b.AesGCM)
	if err != nil {
		loger.Log.Errorf("Active-GetFile-subRoutine: Create session failed! %s", err)
		return nil, err
	}
	defer session.Close()
	_, err = session.SendCommand(command, false, true)
	if err != nil {
		return nil, err
	}

	// 单文件传输线程
	subRoutine := func(inputFile clientComm.GetFile, fileInfo map[string]any, wait *sync.WaitGroup, failFilesChan chan map[string]error) {
		defer wait.Done()

		if inputFile.Offset != 0 {
			b.gfDirect(fileInfo, inputFile, failFilesChan)
		} else {
			b.gfCheck(fileInfo, inputFile, failFilesChan)
		}

	}

	// 开始传输
	//loger.Log.Debugf("Active-GetFile: File %s begins to transfer", inputFiles)
	var wait sync.WaitGroup
	fileCount := len(inputFiles)
	failFileChannel := make(chan map[string]error, fileCount)
	wait.Add(fileCount)
	for _, inputFile := range inputFiles {
		go subRoutine(inputFile, files, &wait, failFileChannel)
	}
	wait.Wait()
	//loger.Log.Debugf("Active-GetFile: File %s transfer completed", inputFiles)
	return
}

// gfDirect 根据偏移量进行传输
func (b *Base) gfDirect(fileInfo map[string]any, inputFile clientComm.GetFile, failFilesChan chan map[string]error) {
	// 初始化传输大小
	dataBlock := int64(config.PacketSize - 8 - b.EncryptionLoss)
	fileMark, ok := fileInfo["fileMark"].(string)
	if !ok {
		loger.Log.Errorf("")
		return
	}
	replyMark, ok := fileInfo["replyMark"].(string)
	if !ok {
		return
	}

	s, err := socket.NewSession(b.TimeChannel, b.DataSocket, nil, replyMark, b.AesGCM)
	defer s.Close()
	if err != nil {
		failFilesChan <- map[string]error{
			inputFile.OutPath: SocketNewSessionErr,
		}
		return
	}

	// 文件输出路径
	outputPath := filepath.Join(inputFile.Space.Path, inputFile.OutPath)

	// 创建文件
	f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0667)
	if err != nil {
		failFilesChan <- map[string]error{
			inputFile.OutPath: FileOpenErr,
		}
		return
	}
	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			loger.Log.Error(err)
			failFilesChan <- map[string]error{
				outputPath: FileCloseErr,
			}
			return
		}
	}(f)

	// 接收文件
	fileBuf := buffer.File{
		TimeChannel: b.TimeChannel,
		FileMark:    fileMark,
		DataBlock:   dataBlock,
	}
	err = fileBuf.FileWrite(f, inputFile.Size, inputFile.Offset, outputPath, inputFile.Date, b.VerifyManage)
	if err != nil {
		if err == errors.New("timeout") {
			loger.Log.Errorf("Active-GetFile-subRoutine: Sync Space %s :Receiving %s file timeout!", inputFile.Space.SpaceName, outputPath)
		} else {
			loger.Log.Errorf("Active-GetFile-subRoutine: Sync Space %s :Unknown error receiving %s file from host %s! %s", inputFile.Space.SpaceName, outputPath, b.Ip, err)
		}
		return
	}

}

// gfCheck 检查并判断文件状态进行传输
func (b *Base) gfCheck(fileInfo map[string]any, inputFile clientComm.GetFile, failFilesChan chan map[string]error) {
	// 初始化传输大小
	dataBlock := int64(config.PacketSize - 8 - b.EncryptionLoss)

	//outPath :=

	fileMark, ok := fileInfo["fileMark"].(string)
	if !ok {
		return
	}
	replyMark, ok := fileInfo["replyMark"].(string)
	if !ok {
		return
	}

	localFileHash, ok := fileInfo["hash"].(string)
	if !ok {
		return
	}
	localFileSize, ok := fileInfo["size"].(int64)
	if !ok {
		return
	}

	// 设置输出路径
	outputPath := filepath.Join(inputFile.Space.Path, inputFile.OutPath)

	// 创建会话接收首次答复
	s, err := socket.NewSession(b.TimeChannel, b.DataSocket, nil, replyMark, b.AesGCM)
	defer s.Close()
	if err != nil {
		failFilesChan <- map[string]error{
			inputFile.OutPath: SocketNewSessionErr,
		}
		return
	}

	reply, err := s.Recv()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			failFilesChan <- map[string]error{
				inputFile.OutPath: SocketNewSessionErr,
			}
			return
		}
		failFilesChan <- map[string]error{
			inputFile.OutPath: err,
		}
		return
	}

	// 处理远程文件状态
	remoteFileDate, ok := reply.Data["date"].(int64)
	if !ok {
		socket.SendStat(s, "Active-GetFile-subRoutine: Missing parameter <date>!")
		return
	}
	remoteFileSize, ok := reply.Data["size"].(int64)
	if !ok {
		socket.SendStat(s, "Active-GetFile-subRoutine: Missing parameter <size>!")
		return
	}
	remoteFileHash, ok := reply.Data["hash"].(string)
	if !ok {
		socket.SendStat(s, "Active-GetFile-subRoutine: Missing parameter <hash>!")
		return
	}

	if remoteFileSize == 0 {
		loger.Log.Errorf("Active-GetFile-subRoutine: File %s failed to retrieve from host %s", outputPath, b.Ip)
		return
	}
	// 初始化文件写入缓冲区
	fileBuf := buffer.File{
		TimeChannel: b.TimeChannel,
		FileMark:    fileMark,
		DataBlock:   dataBlock,
	}
	if remoteFileHash == localFileHash {
		// 正在尝试获取一个同样的文件，没有意义。
	}
	if remoteFileSize > localFileSize && localFileSize != 0 {
		// 1.远程文件大于本地文件，等待对方判断是否为需续写文件；
		// 2.本地文件不存在，需要对方发送文件；

		reply, ok = s.RecvTimeout(int(remoteFileSize/1048576) + 3)
		if !ok {
			loger.Log.Errorf("Active-GetFile-subRoutine: Received file renewal %s reply timeout!", inputFile.OutPath)
			failFilesChan <- map[string]error{
				inputFile.OutPath: SocketTimeoutErr,
			}
			return
		}
		pass, ok := reply.Data["ok"].(bool)
		if !ok || !pass {
			failFilesChan <- map[string]error{
				inputFile.OutPath: ParamsNotExistsErr,
			}
			return
		}

		pathext.MakeDir(filepath.Dir(outputPath))
		f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0667)
		if err != nil {
			failFilesChan <- map[string]error{
				inputFile.OutPath: FileOpenErr,
			}
			return
		}
		defer func(f *os.File) {
			err = f.Close()
			if err != nil {
				loger.Log.Error(err)
				failFilesChan <- map[string]error{
					inputFile.OutPath: FileCloseErr,
				}
				return
			}
		}(f)
		// 创建文件缓冲区(1MB)
		err = fileBuf.FileWrite(f, localFileSize, remoteFileSize, outputPath, remoteFileDate, b.VerifyManage)
		if err == errors.New("timeout") {
			loger.Log.Errorf("Active-GetFile-subRoutine: Sync Space %s :Receiving %s file timeout!", inputFile.Space.SpaceName, outputPath)
		} else {
			loger.Log.Errorf("Active-GetFile-subRoutine: Sync Space %s :Unknown error receiving %s file from host %s! %s", inputFile.Space.SpaceName, outputPath, b.Ip, err)
		}
		return

	} else if remoteFileSize > localFileSize && localFileSize == 0 {
		// 本地文件不存在
		pathext.MakeDir(filepath.Dir(outputPath))
		f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0667)
		if err != nil {
			failFilesChan <- map[string]error{
				inputFile.OutPath: FileOpenErr,
			}
			return
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				loger.Log.Error(err)
				failFilesChan <- map[string]error{
					inputFile.OutPath: FileCloseErr,
				}
				return
			}
		}(f)

		// 创建文件缓冲区(1MB)
		err = fileBuf.FileWrite(f, localFileSize, remoteFileSize, inputFile.OutPath, remoteFileDate, b.VerifyManage)
		if err == errors.New("timeout") {
			loger.Log.Errorf("Active-GetFile-subRoutine: Sync Space %s :Receiving %s file timeout! %s", inputFile.Space.SpaceName, outputPath, err)
			failFilesChan <- map[string]error{
				inputFile.OutPath: SocketTimeoutErr,
			}
			return
		} else {
			loger.Log.Errorf("Active-GetFile-subRoutine: Sync Space %s :Unknown error receiving %s file from host %s! %s", inputFile.Space.SpaceName, outputPath, b.Ip, err)
			failFilesChan <- map[string]error{
				inputFile.OutPath: SocketBufferErr,
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
}
