package config

import (
	"os"
	"path/filepath"
)

// ConfigPath 配置文件存放路径
// var ConfigPath = ".\\data\\config"
// var DevelopConfigPath = path.Join("H:\\Go Project\\exsync-go", "data\\config")

var MainPath = "H:\\Go Project\\exsync-go"

var (
	ConfigPath  = filepath.Join(MainPath, "data\\config")
	LogSavePath = filepath.Join(MainPath, "data")
)

// GetWD 获取当前项目目录
func GetWD() string {
	wd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	return wd
}
