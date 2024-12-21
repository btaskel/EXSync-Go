package muxbuf

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestCopyMarkToSlice(t *testing.T) {
	var uint64v uint64 = 99999999999999
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64v)
	//CopyMarkToSlice(buf, uint64v)
	fmt.Println(buf)
}
