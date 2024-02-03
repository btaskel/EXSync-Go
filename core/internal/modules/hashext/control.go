package hashext

import (
	"bufio"
	"github.com/cespare/xxhash/v2"
	"io"
	"os"
)

// UpdateXXHash 按每次8192个字节更新X次文件的哈希值
func UpdateXXHash(file *os.File, totalBlocks int) (hasher *xxhash.Digest, err error) {
	buffer := make([]byte, 8192)
	xxh := xxhash.New()
	var readBlocks int
	for readBlocks < totalBlocks {
		_, err = bufio.NewReader(file).Read(buffer) // 从渲染器中读取到buffer切片
		if err != nil && err != io.EOF {
			return nil, err
		}
		_, err = xxh.Write(buffer)
		if err != nil {
			return nil, err
		}
		readBlocks += 1
	}
	return xxh, nil
}
