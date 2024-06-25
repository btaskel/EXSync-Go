package compress

import (
	"fmt"
	"github.com/pierrec/lz4"
	"testing"
)

func TestLZ4(t *testing.T) {
	//lz := newLz4(4096)
	originText := []byte{51, 52, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 51}
	//originText := make([]byte, 4096)
	//n, err := rand.Read(originText)
	//if err != nil {
	//	return
	//}
	//originText = originText[:n]

	fmt.Println(originText, len(originText))
	out := make([]byte, 4096)
	block, err := lz4.CompressBlock(originText, out, nil)
	if err != nil {
		return
	}
	fmt.Println(originText[:block], block)

	dst := make([]byte, lz4.CompressBlockBound(block))
	n, err := lz4.UncompressBlock(out[0:block], dst)
	if err != nil {
		return
	}
	fmt.Println(dst[:n], len(dst[:n]))

}
