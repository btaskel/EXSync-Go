package pathext

import (
	"fmt"
	"os"
)

// MakeDir 如果一个文件不存在，则创建它的路径
func MakeDir(dirPath string) (ok bool) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		errDir := os.MkdirAll(dirPath, 0755)
		if errDir != nil {
			fmt.Println("Error creating directory")
			fmt.Println(errDir)
			return
		}
	}
	return true
}
