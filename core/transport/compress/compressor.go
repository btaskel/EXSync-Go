package compress

import (
	"errors"
	"strings"
)

const (
	Lz4 = "lz4" // Best
)

var (
	// ErrUnsupportedCompressionMethod 不受支持的压缩算法
	ErrUnsupportedCompressionMethod = errors.New("unsupported compression method")
)

type CompressorInfo struct {
	lossLen     int
	newCompress func(compressLen, loss int) Compress
}

var compressorMethod = map[string]CompressorInfo{
	Lz4: {lossLen: 1, newCompress: newLz4},
}

type Compress interface {
	CompressData(src, dst []byte) (int, error)
	UnCompressData(src, dst []byte) (int, error)
	GetDstBuf() []byte
	GetLoss() int
}

// NewCompress 初始化网络压缩器
func NewCompress(method string, compressLen int) (Compress, int, error) {
	if compressMethod, ok := compressorMethod[strings.ToLower(method)]; ok {
		return compressMethod.newCompress(compressLen, compressMethod.lossLen), compressMethod.lossLen, nil
	}
	return nil, 0, ErrUnsupportedCompressionMethod
}
