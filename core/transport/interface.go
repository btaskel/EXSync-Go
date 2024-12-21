package transport

import (
	"context"
	"github.com/quic-go/quic-go"
	"io"
	"net"
)

type Listener interface {
	Addr() net.Addr
	io.Closer
	Accept(ctx context.Context) (quic.Connection, error)
}

type Conn interface {
	AcceptStream(context.Context) (quic.Stream, error)
	// OpenStream opens a new bidirectional QUIC stream.
	// There is no signaling to the peer about new streams:
	// The peer can only accept the stream after data has been sent on the stream,
	// or the stream has been reset or closed.
	// When reaching the peer's stream limit, it is not possible to open a new stream until the
	// peer raises the stream limit. In that case, a StreamLimitReachedError is returned.
	OpenStream() (quic.Stream, error)
	// OpenStreamSync opens a new bidirectional QUIC stream.
	// It blocks until a new stream can be opened.
	// There is no signaling to the peer about new streams:
	// The peer can only accept the stream after data has been sent on the stream,
	// or the stream has been reset or closed.
	OpenStreamSync(context.Context) (quic.Stream, error)
	// LocalAddr returns the local address.
	LocalAddr() net.Addr
	// RemoteAddr returns the address of the peer.
	RemoteAddr() net.Addr
	// CloseWithError closes the connection with an error.
	// The error string will be sent to the peer.
	CloseWithError(quic.ApplicationErrorCode, string) error
	// Context returns a context that is cancelled when the connection is closed.
	// The cancellation cause is set to the error that caused the connection to
	// close, or `context.Canceled` in case the listener is closed first.
	Context() context.Context
	// ConnectionState returns basic details about the QUIC connection.
	// In TCP, it returns nothing
	// In TCP over TLS, it returns the TLS status
	// Warning: This API should not be considered stable and might change soon.
	ConnectionState() quic.ConnectionState
}

// Stream with Cipher
type Stream interface {
	// GetBuf 分配一个缓冲区内存, 并返回可以写入的起始索引
	GetBuf() ([]byte, int)
	// Reader 读取填充缓冲区, 返回一个读取末尾索引.
	// 该缓冲区没有必要被填充满, 如果 Stream 被关闭则会返回ErrClosed
	io.Reader
	// Writer ...
	io.Writer
	// Closer 关闭连接以及所有 Stream
	io.Closer
}
