package hashext

import (
	"fmt"
	"github.com/cespare/xxhash/v2"
	"os"
)

// GetFileBlockHash 获取文件的分块（1MB）哈希值
// path 文件路径
func GetFileBlockHash(path string) (blockHash, totalHash string, err error) {
	block := 1048576
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	buf := make([]byte, block)

	for {
		hasher := xxhash.New()
		n, err := f.Read(buf)
		if n == 0 {
			break
		}
		_, err = hasher.Write(buf[:n])
		if err != nil {
			return "", "", err
		}
		blockHash += fmt.Sprintf("%X", hasher.Sum(nil))
		if err != nil {
			return "", "", err
		}
	}

}
