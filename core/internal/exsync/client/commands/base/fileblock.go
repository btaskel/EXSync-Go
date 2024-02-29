package base

import (
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/internal/modules/socket"
	loger "EXSync/core/log"
	"EXSync/core/option/exsync/comm"
)

type FileBlock struct {
	RelPath   string
	OutPath   string
	TotalSize int64
	Offset    []int64
}

type FileInfo struct {
	Total  int64   // 总量
	Offset []int64 // 数据偏移量指针
	Mark   string
}

// GetFileBlock 分块获取远程文件
// 根据本地数据库与远程数据库的比较, 本地直接获取相应远程文件块, 而不需要再次交流.
func (b *Base) GetFileBlock(files []FileBlock) {
	fileInfo := make(map[string]FileInfo, len(files))

	for _, file := range files {
		fileInfo[file.RelPath] = FileInfo{
			Total:  file.TotalSize,
			Offset: file.Offset,
			Mark:   hashext.GetRandomStr(6),
		}
	}

	command := comm.Command{
		Command: "data",
		Type:    "fileBlock",
		Method:  "get",
		Data: map[string]any{
			"fileInfo": fileInfo,
		},
	}

	reply := hashext.GetRandomStr(6)
	s, err := socket.NewSession(b.TimeChannel, b.DataSocket, b.CommandSocket, reply, b.AesGCM)
	if err != nil {
		loger.Log.Errorf("")
		return
	}
	_, err = s.SendCommand(command, false, true)
	if err != nil {
		return
	}

	subRoutine := func(relPath string, info FileInfo) {

	}

	for relPath, info := range fileInfo {

	}

}
