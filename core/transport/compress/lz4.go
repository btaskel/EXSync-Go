package compress

import (
	"github.com/pierrec/lz4"
)

type Lz4Compress struct {
	srcSize int
	loss    int
}

// newLz4 4096 - 2 DataLen - 4 StreamMarkLen - N Cipher
// compressLen: 需要压缩的数据长度
func newLz4(compressLen, loss int) Compress {
	return &Lz4Compress{srcSize: compressLen, loss: loss}
}

func (c *Lz4Compress) CompressData(src, dst []byte) (int, error) {
	return lz4.CompressBlock(src, dst, nil)
}

func (c *Lz4Compress) UnCompressData(src, dst []byte) (int, error) {
	return lz4.UncompressBlock(src, dst)
}

// GetDstBuf 获取一个解压大小的切片
func (c *Lz4Compress) GetDstBuf() []byte {
	return make([]byte, lz4.CompressBlockBound(c.srcSize))
}

func (c *Lz4Compress) GetLoss() int {
	return c.loss
}
