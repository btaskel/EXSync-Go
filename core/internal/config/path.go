package config

import (
	"os"
	"path/filepath"
)

// ConfigSavePath 配置文件存放路径
// var ConfigSavePath = ".\\data\\config"
// var DevelopConfigPath = path.Join("H:\\Go Project\\exsync-go", "data\\config")

var (
	SpaceMainPath = ".sync" // exsync同步空间保存的主路径

	SpaceInfoPath = filepath.Join(SpaceMainPath, "db") // exsync数据存放路径
)

var (
	MainPath = "H:\\Go Project\\exsync-go" // exsync工作目录

	ConfigSavePath = filepath.Join(MainPath, "data\\config") // 配置文件保存路径
	LogSavePath    = filepath.Join(MainPath, "data\\logs")   // 日志保存路径

	IndexSavePath = filepath.Join(MainPath, "data\\index")
)

// GetWD 获取当前项目所在目录
func GetWD() string {
	wd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	return wd
}
