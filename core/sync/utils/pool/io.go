package pool

import (
	"EXSync/core/internal/exsync/client"
	configOption "EXSync/core/option/config"
	serverOption "EXSync/core/option/exsync/manage"
	"context"
)

type file struct {
	RelPath   string
	OutPath   string
	FileHash  string
	FileSize  int64
	FileDate  int64
	Hosts     []string
	Space     *configOption.UdDict
	Offset    int64 // 拉取文件时根据偏移量来获取文件部分
	TotalSize int64 // 拉取总量
}

// HostStress 主机压力映射
type HostStress = *struct {
	tasks  chan file
	total  int64
	client *client.Client
}

type Pool struct {
	threads             int // 每个主机最大并发数
	queueNum            int // 每个主机最大等待队列 default: 64
	waitQueue           chan file
	hostStress          map[Host]HostStress
	activeConnectManage map[string]serverOption.ActiveConnectManage
	ctx                 context.Context
}

func NewFilePool(ctx context.Context, ActiveConnectManage map[string]serverOption.ActiveConnectManage) *Pool {
	pool := &Pool{
		waitQueue:           make(chan file),
		hostStress:          make(map[Host]HostStress),
		activeConnectManage: ActiveConnectManage,
		ctx:                 ctx,
	}
	pool.ScanHost()
	go pool.checkTask()

	return pool
}

// Add 增加任务
func (p *Pool) Add(file file) {
	p.waitQueue <- file
}

// ScanHost 扫描主机并增加至池
func (p *Pool) ScanHost() {
	// 更新hostStress
	for hostName := range p.activeConnectManage {
		if _, ok := p.hostStress[hostName]; ok {
			continue
		} else {
			go p.addHost(hostName)
		}
	}
}

// Close 关闭池
func (p *Pool) Close() {

}
