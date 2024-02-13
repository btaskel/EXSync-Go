package config

import (
	"os"
	"path"
)

// ConfigPath 配置文件存放路径
// var ConfigPath = ".\\data\\config"
// var DevelopConfigPath = path.Join("H:\\Go Project\\exsync-go", "data\\config")
var ConfigPath = path.Join("H:\\Go Project\\exsync-go", "data\\config")

// 获取当前项目目录
func GetWD() string {
	wd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	return wd
}
