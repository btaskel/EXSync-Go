package sync

import "EXSync/core/internal/exsync/server"

type Sync struct {
	Server *server.Server
}

func (s *Sync) NewSync() {
	// 启动运行服务
	ser := server.NewServer()
	ser.Run()

}
