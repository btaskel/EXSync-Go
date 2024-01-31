package base

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/exsync/server/index"
	"EXSync/core/internal/modules/socket"
	"EXSync/core/option"
	"github.com/sirupsen/logrus"
	"path"
)

func (c *Base) GetFile() {

}

// PostFile 客户端发送文件至服务端
func (c *Base) PostFile(data map[string]any, mark string) {
	remoteSpace := data["remote_space_name"].(string)
	localSpace, ok := config.UserData[remoteSpace]
	if !ok {
		return
	}
	remoteSpaceName, ok := data["spacename"].(string)
	if !ok {
		return
	}
	remoteFileRelPath, ok := data["file_path"].(string)
	if !ok {
		return
	}
	remoteFileSize, ok := data["file_size"].(int)
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

	remoteFileAbsPath := path.Join(localSpace.Path, remoteFileRelPath)
	if len(remoteFileHash) != 32 {
		status = "File hash too long!"
	}

	// 获取db
	spaceDb, err := index.Index.GetIndex(path.Join(localSpace.Path, ".\\sync\\info\\files.db"))
	if err != nil {
		logrus.Errorf("Host %s failed to open the index database for %s!", c.Ip, remoteSpace)
	}

	var file option.Index
	var command option.Command
	spaceDb.Where("Path = ?", remoteFileRelPath).First(&file)
	//spaceIndex.First(&indexOption, "code = ?", "D42")

	//if result.Error != nil {
	//	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
	//		size = 0
	//	} else {
	//		logrus.Errorf("")
	//	}
	//}
	if file.Path == "" {
		command = option.Command{
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
		command = option.Command{
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
	session, err := socket.NewSession(c.TimeChannel, c.DataSocket, nil, mark, c.AesGCM)
	if err != nil {
		return
	}

	result, err := session.SendCommand(command, true, true)
	if err != nil {
		return
	}

	dataBlock := config.PacketSize - 8 - c.EncryptionLoss // fileMark 8 + GCM-tag 16 + GCM-nonce 12

}
