package syncloop

import (
	configOption "EXSync/core/option/config"
	"EXSync/core/sync/methods/monitor"
)

// 0 : 监控
// 1 : 双端
// 2 : 多端

type methods struct {
	monitor *monitor.Monitor
}

type Loop struct {
	methods
	notice   chan int
	userData configOption.UdDict
}
