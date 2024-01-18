package main

import (
	"log"
	"os"
)

func main() {
	// 打开文件以读写模式
	file, err := os.OpenFile("file.txt", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 数据块
	dataChunks := [][]byte{
		[]byte("Hello, "),
		[]byte("world!"),
		// 更多数据块...
	}

	// 写入数据块
	pos := int64(0)
	for _, chunk := range dataChunks {
		n, err := file.WriteAt(chunk, pos)
		if err != nil {
			log.Fatal(err)
		}
		pos += int64(n)

		// 确保数据块被写入磁盘
		if err := file.Sync(); err != nil {
			log.Fatal(err)
		}
	}
}
