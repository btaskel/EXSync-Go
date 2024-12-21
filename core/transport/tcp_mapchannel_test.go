package transport

import (
	"EXSync/core/transport/muxbuf"
	"fmt"
	"testing"
)

// TestNewTimeChannel pass
func TestNewTimeChannel(t *testing.T) {
	mark := muxbuf.Mark(1)
	tc := newTimeChannel(4090)
	defer tc.close()
	err := tc.createRecv(mark)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tc.del(mark)

	buf := []byte{50, 51, 52, 53}
	muxBuf, ok := tc.getMuxBuf(mark)
	if !ok {
		fmt.Println("mark不存在")
		return
	}
	err = tc.push(muxBuf, &buf)
	if err != nil {
		return
	}
	buf2 := make([]byte, 4096)
	muxBuf, ok = tc.getMuxBuf(mark)
	if !ok {
		fmt.Println("mark不存在")
		return
	}
	n, err := tc.pop(muxBuf, &buf2)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(buf2[:n])
}
