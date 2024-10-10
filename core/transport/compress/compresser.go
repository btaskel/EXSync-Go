package compress

import (
	"errors"
	"strings"
)

const (
	Lz4  = "lz4" // Best
	Gzip = "gzip"
)

var (
	// ErrUnsupportedCompressionMethod 不受支持的压缩算法
	ErrUnsupportedCompressionMethod = errors.New("unsupported compression method")
)

var methodMap = map[string]struct {
	lossLen     int
	newCompress func(compressLen int) Compress
}{
	Lz4:  {lossLen: 1, newCompress: newLz4},
	Gzip: {lossLen: 1},
}

type Compress interface {
	CompressData(src, dst []byte) (int, error)
	UnCompressData(src, dst []byte) (int, error)
	GetDstBuf() []byte
}

// NewCompress 初始化网络压缩器
func NewCompress(method string, compressLen int) (Compress, int, error) {
	if compressMethod, ok := methodMap[strings.ToLower(method)]; ok {

		return compressMethod.newCompress(compressLen), compressMethod.lossLen, nil
	}
	return nil, 0, ErrUnsupportedCompressionMethod
}
