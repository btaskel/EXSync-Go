package muxbuf

import (
	"errors"
	"fmt"
)

var (
	ErrMarkNotExist = errors.New("mark not exist")
	ErrMarkExist    = errors.New("mark exist")
)

// Mark 作为MapChannel的MapChan键使用
type Mark [4]byte

// SliceToMark 取出前4个元素,将数组转换为mapChannel的Mark类型
func SliceToMark(slice []byte) Mark {
	var arr = Mark{}
	for i := 0; i < 4; i++ {
		fmt.Println(arr, slice)
		arr[i] = slice[i]
	}
	return arr
}

// CopyMarkToSlice Mark复制到切片中
func CopyMarkToSlice(dst []byte, src Mark) {
	for i := 0; i < 4; i++ {
		dst[i] = src[i]
	}
}

// MarkToUint32 等效于 binary.BigEndian.Uint32
func MarkToUint32(b Mark) uint32 {
	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
}

// PutUint32 等效于 binary.BigEndian.PutUint32
func PutUint32(v uint32) (b Mark) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
	return b
}
