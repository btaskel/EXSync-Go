package pool

import (
	configOption "EXSync/core/option/config"
	serverOption "EXSync/core/option/exsync/server"
)

type StressManage struct {
	threads  int // 每个主机最大并发数
	queueNum int // 每个主机最大等待队列 default: 64
}

type file struct {
	RelPath   string
	OutPath   string
	FileHash  string
	FileSize  int64
	FileDate  int64
	Hosts     []string
	Space     *configOption.UdDict
	Offset    []int64 // 拉取文件时根据偏移量来获取文件部分, 默认每次拉取偏移之后的1MB数据
	TotalSize int64   // 拉取总量
}

type Pool struct {
	StressManage
	waitQueue  chan file
	dynamic    chan file
	hostStress map[Host]*struct {
		tasks chan file
		total int64
	}
	activeConnectManage map[string]serverOption.ActiveConnectManage
}

func NewFilePool(ActiveConnectManage map[string]serverOption.ActiveConnectManage) *Pool {
	pool := &Pool{
		waitQueue: make(chan file),
		dynamic:   make(chan file),
		hostStress: make(map[Host]*struct {
			tasks chan file
			total int64
		}),
		activeConnectManage: ActiveConnectManage,
	}

	go pool.checkTask()

	return pool
}

// Add 增加任务
func (p *Pool) Add(file file) {
	if len(file.Hosts) == 1 {
		p.waitQueue <- file
	} else {
		p.dynamic <- file
	}
}
