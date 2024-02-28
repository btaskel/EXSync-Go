package pool

import (
	configOption "EXSync/core/option/config"
	serverOption "EXSync/core/option/exsync/server"
)

type StressManage struct {
	threads int
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

func (p *Pool) checkTask() {
	for {
		select {
		case fileTask := <-p.waitQueue:
			// 更新hostStress
			for hostName := range p.activeConnectManage {
				if _, ok := p.hostStress[hostName]; ok {
					continue
				} else {
					go p.initHost(hostName)
				}
			}

			// 获取当前压力最小的主机
			minV := int64(^uint64(0) >> 1)
			host := ""
			for k, v := range p.hostStress {
				if v.total < minV {
					minV = v.total
					host = k
				}
			}

			//添加任务
			p.hostStress[host].tasks <- fileTask
			p.hostStress[host].total += fileTask.TotalSize
		}
	}
}

func (p *Pool) initHost(hostName string) {
	p.hostStress[hostName] = &struct {
		tasks chan file
		total int64
	}{tasks: make(chan file), total: 0}

	for {
		select {
		case t := <-p.hostStress[hostName].tasks:
			// todo: getBlock 未完待续——（bushi
		}
	}
}

// Add 增加任务
func (p *Pool) Add(file file) {
	if len(file.Hosts) == 1 {
		p.waitQueue <- file
	} else {
		p.dynamic <- file
	}
}
