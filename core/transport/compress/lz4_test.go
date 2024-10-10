package compress

import (
	"fmt"
	"github.com/pierrec/lz4"
	"testing"
)

func TestLZ4(t *testing.T) {
	originText := []byte{51, 52, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 51}
	output := make([]byte, 4096)
	n, err := lz4.CompressBlock(originText, output[4:], nil)
	if err != nil {
		return
	}
	fmt.Println("lz4-compressData", originText[:n], n)

	dst := make([]byte, lz4.CompressBlockBound(n))
	n, err = lz4.UncompressBlock(output[:n], dst)
	if err != nil {
		return
	}
	fmt.Println(dst[:n], len(dst[:n]))
}

// TestNewCompress lz4
func TestNewCompress(t *testing.T) {
	compress, n, err := NewCompress("lz4", 4090)
	if err != nil {
		return
	}
	originText := "测试文字"
	compressSlice := make([]byte, 4096)
	n, err = compress.CompressData([]byte(originText), compressSlice)
	if err != nil {
		return
	}
	fmt.Println("compressData: ", compressSlice[:n], n)

	unCompressSlice := make([]byte, 4096)
	n, err = compress.UnCompressData(compressSlice[:n], unCompressSlice)
	if err != nil {
		return
	}
	fmt.Println("uncompressedData: ", unCompressSlice[:n], n)
	fmt.Println("uncompressedData-string: ", string(unCompressSlice[:n]), n)
}
