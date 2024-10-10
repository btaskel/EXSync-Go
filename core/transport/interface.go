package transport

import (
	"context"
)

type Conn interface {
	AcceptStream(context.Context) (Stream, error)
	// OpenStream 打开一个流，如果等待超时则返回EOF
	OpenStream() (Stream, error)
	// OpenStreamSync 打开
	OpenStreamSync(context.Context) (Stream, error)

	Close() error
}

// Stream with Cipher
type Stream interface {
	// GetBuf 分配一个缓冲区内存, 并返回可以写入的起始索引
	GetBuf() ([]byte, int)
	// Read 读取填充缓冲区, 返回一个读取末尾索引.
	// 该缓冲区没有必要被填充满, 如果 Stream 被关闭则会返回ErrClosed
	Read(b []byte) (int, error)
	// Write ...
	Write(b []byte) (int, error)
	// Close 关闭连接以及所有 Stream
	Close() error
}