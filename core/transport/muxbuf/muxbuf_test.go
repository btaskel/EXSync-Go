package muxbuf

import (
	"fmt"
	"testing"
)

func TestNewMuxBuf(t *testing.T) {
	var slicePoint *[]byte
	s1 := []byte{60, 61, 62, 63, 64}

	mb := NewMuxBuf(4, 4096)
	slicePoint, _ = mb.PickFree()
	*slicePoint = (*slicePoint)[:len(s1)]
	fmt.Println(*slicePoint)
	copy(*slicePoint, s1)

	mb.PutUsed(slicePoint)

	slicePoint, _ = mb.PickUsed()

	fmt.Println(slicePoint)
}

func TestMuxBuf_PickFree(t *testing.T) {
	channel := make(chan *[]byte)
	close(channel)
	chann := <-channel
	chann = <-channel
	if chann == nil {
		fmt.Println("chan is nil")
	} else {
		fmt.Println(chann)
	}
}
