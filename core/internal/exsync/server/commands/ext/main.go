package ext

import (
	"EXSync/core/internal/exsync/server/commands/base"
	"EXSync/core/option/exsync/comm"
)

type CommandSet struct {
	base.Base
}

// NewCommandSet 创建扩展指令集对象
func NewCommandSet(b base.Base) (*CommandSet, error) {
	return &CommandSet{b}, nil
}

// MatchCommand 匹配命令到相应的函数
func (c *CommandSet) MatchCommand(command comm.Command) {
	switch command.Command {
	case "comm":
		switch command.Type {
		case "verifyConnect":
		case "command":
		case "shell":
		}
	case "data":
		switch command.Type {
		case "file":
			switch command.Method {
			case "get":
				c.GetFile(command.Data)
			case "post":
			}
		case "folder":
			switch command.Method {
			case "get":
			case "post":
			}
		}
	case "index":
		switch command.Method {
		case "get":
		}
	}
}
