package methods

import configOption "EXSync/core/option/config"

// NewSpaceProcess 获取所有同步空间的信息

type SpaceProcess struct {
	SyncSpace configOption.UdDict
}

func NewSpaceProcess(syncSpace configOption.UdDict) *SpaceProcess {
	process := SpaceProcess{SyncSpace: syncSpace}
	return &process
}
