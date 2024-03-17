package sync

import (
	"EXSync/core/internal/exsync/server"
)

type Sync struct {
	Server    *server.Server
	spaceLock map[string]struct{}
}

func (s *Sync) NewSync() {
	// 运行服务
	ser := server.NewServer()
	ser.Run()

}
