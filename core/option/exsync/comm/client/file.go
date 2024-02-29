package clientComm

import configOption "EXSync/core/option/config"

// GetFile 输入
type GetFile struct {
	RelPath string
	OutPath string
	Size    int64 // 当前文件大小
	Date    int64 // 当前文件unix时间, 如果offset不为空则为本地保持目标时间
	Hash    string
	Space   *configOption.UdDict
	Offset  int64 // 获取部分的偏移量
}
