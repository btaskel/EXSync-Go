package ext

import (
	"EXSync/core/internal/exsync/client/commands/base"
)

type CommandSet struct {
	base.Base
}

func NewCommandSet(clientCommBase base.Base) *CommandSet {
	return &CommandSet{clientCommBase}
}
