package base

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/pathext"
	"EXSync/core/internal/modules/socket"
	loger "EXSync/core/log"
	configOption "EXSync/core/option/config"
	"EXSync/core/option/exsync/comm"
	clientComm "EXSync/core/option/exsync/comm/client"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type IndexFile struct {
	path string
}

// Close 清理远程索引缓存
func (i *IndexFile) Close() {
	_ = os.Remove(i.path)
}

// GetIndex 获取当前
func (b *Base) GetIndex(space configOption.UdDict) (*IndexFile, error) {
	if !pathext.MakeDir(filepath.Join(space.Path, config.SpaceInfoPath)) {
		return nil, errors.New("PathNotExist")
	}

	// 判断远程文件索引数据库是否为最新
	indexFilePath := filepath.Join(space.Path, config.SpaceInfoPath, "sync.db")
	file, err := os.Stat(indexFilePath)
	localFileDate := file.ModTime().Unix()
	if err != nil {
		return nil, err
	}
	replyMark := hashext.GetRandomStr(6)
	command := comm.Command{
		Command: "data",
		Type:    "index",
		Method:  "get",
		Data: map[string]any{
			"spaceName": space.SpaceName,
			"replyMark": replyMark,
		},
	}
	s, err := socket.NewSession(b.TimeChannel, b.DataSocket, b.CommandSocket, replyMark, b.AesGCM)
	if err != nil {
		loger.Log.Errorf("Active-GetIndex: Create session failed! %s", err)
		return nil, err
	}
	reply, err := s.SendCommand(command, true, true)
	if err != nil {
		return nil, err
	}

	// 获取远程参数
	data, ok := reply["data"].(map[string]any)
	if !ok {
		socket.SendStat(s, "Active-GetIndex: Missing parameter <data>!")
		return nil, errors.New("missing parameter")
	}

	remoteFileDate, ok := data["date"].(int64)
	if !ok || remoteFileDate == 0 {
		socket.SendStat(s, "Active-GetIndex: Missing parameter <date>!")
		return nil, errors.New("missing parameter")
	}

	if remoteFileDate-b.VerifyManage.Offset > localFileDate {
		// 创建数据库缓存文件夹
		saveFolderName := fmt.Sprintf("%s%s%d", b.Ip, space.SpaceName, remoteFileDate)
		saveFolderNameHash := hashext.GetSha256(saveFolderName)
		savePath := filepath.Join(config.IndexSavePath, saveFolderNameHash)
		pathext.MakeDir(savePath)
		remoteIndexFilePath := filepath.Join(savePath, "sync.db")

		getFile := []clientComm.GetFile{{
			RelPath: filepath.Join(config.SpaceInfoPath, "sync.db"),
			OutPath: remoteIndexFilePath,
			Space:   &space,
		}}
		_, err = b.GetFile(getFile)
		if err == nil {
			index := IndexFile{path: remoteIndexFilePath}
			loger.Log.Debugf("Active-GetIndex: Successfully obtained the index of SyncSpace %s from host %s", space.SpaceName, b.Ip)
			return &index, err
		} else {
			loger.Log.Errorf("Active-GetIndex: Unable to obtain the index of SyncSpace %s from host %s! %s", space.SpaceName, b.Ip, err)
			return nil, err
		}
	} else {
		loger.Log.Debugf("Active-GetIndex: Retrieve the index of SyncSpace %s from host %s, but there is no update.", space.SpaceName, b.Ip)
		return nil, errors.New("NoUpdate")
	}
}

func PostIndex() {

}
