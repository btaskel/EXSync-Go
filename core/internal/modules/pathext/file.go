package pathext

import (
	"fmt"
	"os"
	"path/filepath"
)

// MakeDir 如果一个文件不存在，则创建它的路径
func MakeDir(filePath string) (ok bool) {
	dirPath := filepath.Dir(filePath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		errDir := os.MkdirAll(dirPath, 0775)
		if errDir != nil {
			fmt.Println("Error creating directory")
			fmt.Println(errDir)
			return
		}
	}
	return true
}
