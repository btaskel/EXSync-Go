package base

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/socket"
	"EXSync/core/option"
	"EXSync/core/option/server/comm"
	"encoding/json"
	"gorm.io/gorm"
	"sync"
)

// GetFile 根据相对路径列表在Space里搜索相应的文件，并获取到本机。
// relPaths的数量将决定以多少并发获取文件
func (c *Base) GetFile(relPaths []string, db *gorm.DB, space comm.UdDict) (ok bool, err error) {
	// 初始化GetFile
	var (
		hashList      []string
		sizeList      []int64
		replyMarkList []string
		fileMarkList  []string
	)
	for _, relPath := range relPaths {
		var file option.Index
		fileMark := hashext.GetRandomStr(8)
		replyMark := hashext.GetRandomStr(8)
		db.Where("path = ?", relPath).First(&file)
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
			"pathList":     relPaths,
			"fileHashList": hashList,
			"fileSizeList": sizeList,
			"fileMarkList": fileMarkList,
			"spaceName":    space.SpaceName,
		},
	}
	session, err := socket.NewSession(c.TimeChannel, nil, c.CommandSocket, hashext.GetRandomStr(8), c.AesGCM)
	if err != nil {
		return false, err
	}
	_, err = session.SendCommand(command, false, true)
	if err != nil {
		return false, err
	}

	subRoutine := func(relPath, localFileHash string, localFileSize int64, fileMark, replyMark string, wait *sync.WaitGroup, channel chan string) {
		defer wait.Done()
		defer func() {
			channel <- relPath
		}()
		// 创建会话接收首次答复
		session, err = socket.NewSession(c.TimeChannel, c.DataSocket, nil, replyMark, c.AesGCM)
		if err != nil {
			return
		}
		data, ok := session.Recv()
		if !ok {
			return
		}
		var reply comm.Command
		err = json.Unmarshal(data, &reply)
		if err != nil {
			return
		}
		//
	}

	// 初始化传输
	dataBlock := int64(config.PacketSize - 8 - c.EncryptionLoss)

	// 开始传输
	var wait sync.WaitGroup
	fileCount := len(relPaths)
	wait.Add(fileCount)
	channel := make(chan string, fileCount)
	for i := 0; i < fileCount; i++ {
		go subRoutine(relPaths[i], hashList[i], sizeList[i], fileMarkList[i], replyMarkList[i], &wait, channel)
	}
	wait.Wait()

}

func (c *Base) PostFile() {

}
