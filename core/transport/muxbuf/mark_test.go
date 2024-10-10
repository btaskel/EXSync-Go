package muxbuf

import (
	"fmt"
	"testing"
)

func TestPutUint32(t *testing.T) {
	arr := PutUint32(32)
	fmt.Println(arr)
}

func TestMarkToUint32(t *testing.T) {
	arr := PutUint32(32)
	fmt.Println(MarkToUint32(arr))
}
