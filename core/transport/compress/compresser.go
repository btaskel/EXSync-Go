package compress

import (
	"errors"
	"strings"
)

const (
	Lz4 = "lz4"
)

var methodMap = map[string]struct {
	lossLen     int
	newCompress func(compressLen int) Compress
}{
	Lz4: {lossLen: 1, newCompress: newLz4},
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
	return nil, 0, errors.New("unsupported compression method")
}
