package serverOption

import (
	"EXSync/core/internal/exsync/client"
	"EXSync/core/internal/exsync/server/commands"
	"time"
)

// ActiveConnectManage 主动连接管理
type ActiveConnectManage struct {
	ID         string
	CreateTime time.Time
	Client     *client.Client
}

// PassiveConnectManage 被动连接管理
type PassiveConnectManage struct {
	ID             string
	CreateTime     time.Time
	CommandProcess *commands.CommandProcess
}
