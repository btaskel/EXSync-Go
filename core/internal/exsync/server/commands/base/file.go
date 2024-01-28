package base

import (
	"EXSync/core/internal/config"
)

func (c *CommandSet) getFile() {

}

// postFile 客户端发送文件至服务端
func (c *CommandSet) postFile(data map[string]any, mark string) {
	localSpace, ok := config.UserData["remote_space_name"]
	if !ok {
		return
	}
	remoteSpaceName, ok := data["spacename"].(string)
	if !ok {
		return
	}
	remoteFileRelativePath, ok := data["file_path"].(string)
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

	//status := "ok"

}
