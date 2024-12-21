package muxbuf

import (
	"errors"
)

var (
	ErrMarkNotExist = errors.New("mark not exist")
	ErrMarkExist    = errors.New("mark exist")
)

const MarkLen = 8

// Mark 作为MapChannel的MapChan键使用
// type Mark [MarkLen]byte
type Mark = uint64
