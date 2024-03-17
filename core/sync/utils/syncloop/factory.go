package syncloop

import (
	"EXSync/core/internal/exsync/server"
	configOption "EXSync/core/option/config"
	"EXSync/core/sync/methods/monitor"
)

func NewSyncLoop(server *server.Server, userData configOption.UdDict) {
	l := Loop{
		methods:  methods{},
		notice:   make(chan int),
		userData: userData,
	}
	l.syncLoop()
}

func (s *Loop) initMethods() {
	s.monitor = monitor.NewMonitor()
}
