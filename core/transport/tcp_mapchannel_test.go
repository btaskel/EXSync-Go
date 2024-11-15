package transport

import (
	"EXSync/core/transport/muxbuf"
	"fmt"
	"io"
	"testing"
)

// TestNewTimeChannel pass
func TestNewTimeChannel(t *testing.T) {
	mark := muxbuf.Mark{0, 0, 0, 1}
	tc := NewTimeChannel(4090)
	defer tc.Close(io.EOF)
	err := tc.CreateRecv(mark)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tc.Del(mark)

	buf := []byte{50, 51, 52, 53}
	muxBuf, ok := tc.GetMuxBuf(mark)
	if !ok {
		fmt.Println("mark不存在")
		return
	}
	err = tc.Push(muxBuf, &buf)
	if err != nil {
		return
	}
	buf2 := make([]byte, 4096)
	muxBuf, ok = tc.GetMuxBuf(mark)
	if !ok {
		fmt.Println("mark不存在")
		return
	}
	n, err := tc.Pop(muxBuf, &buf2)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(buf2[:n])
}
