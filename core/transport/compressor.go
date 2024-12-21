package transport

import "EXSync/core/transport/compress"

func RegisterCompressor(method string, compressorInfo compress.CompressorInfo) error {
	return compress.Register(method, compressorInfo)
}
