package transport

import (
	"context"
)

type Conn interface {
	AcceptStream(context.Context) (Stream, error)

	OpenStream() (Stream, error)
	OpenStreamSync(context.Context) (Stream, error)

	Close() error
}

// Stream with Cipher
type Stream interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}
