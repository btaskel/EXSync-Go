package methods

import (
	configOption "EXSync/core/option/config"
	"time"
)

// NewSpaceProcess 获取所有同步空间的信息
func NewSpaceProcess(syncSpace configOption.UdDict) *SpaceProcess {
	process := SpaceProcess{SyncSpace: syncSpace}
	return &process
}

func (s *SpaceProcess) queue() {

	for {

		time.Sleep(time.Duration(s.SyncSpace.Interval) * time.Second)
	}
}
