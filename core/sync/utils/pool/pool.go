// 文件同步池:
// 在与某一台主机的同步空间进行同步时，为了尽可能地进行负载均衡, exsync会使用文件池动态调整文件拉取的目标主机。
//
// 当exsync对一个同步空间进行同步操作时, 会先拉取与之连接的N个主机的同一同步空间的文件索引。
// 之后将需要同步的文件存储至一个队列, 在理想情况下这个队列中的文件会按块均匀地分散到N个主机,
// 并从中请求相应数据的传输。如果请求某一主机失败或者遇到其它错误，将会从其它主机拉取，这个过
// 程会尽可能地保持每个主机相同的压力。

// 1.获取所有主机的index
// 2.分析index, 得出需要同步的文件
// 3.分析需要同步的文件, 得出需要动态处理的文件
// 4.运行

package pool

import (
	"EXSync/core/internal/exsync/client"
	"time"
)

// 检查任务所需的主机, 并筛选出压力最小的主机与之进行传输
func (p *Pool) checkTask() {
	for {
		select {
		case fileTask := <-p.waitQueue:
			// 获取该任务准许的主机
			var allowHost map[Host]HostStress
			for _, taskHostName := range fileTask.Hosts {
				for hostName, activeConnect := range p.hostStress {
					if hostName == taskHostName {
						allowHost[hostName] = activeConnect
					}
				}
			}

			// 获取当前压力最小的主机
			minV := int64(^uint64(0) >> 1)
			host := ""
			for k, v := range allowHost {
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

// addHost 增加主机, 并使该主机从文件同步池中持续接收任务
func (p *Pool) addHost(hostAddr string) {
	c := p.activeConnectManage[hostAddr].Client

	p.hostStress[hostAddr] = &struct {
		tasks  chan file
		total  int64
		client *client.Client
	}{tasks: make(chan file, p.queueNum), total: 0, client: c}

	// 任务处理: 每次从队列取十个任务, 如果超时则将现有的任务直接提交
	var files map[string]file
	for {
		select {
		case t, ok := <-p.hostStress[hostAddr].tasks:
			if !ok {
				return
			}

			files[t.RelPath] = t
			if len(files) == 10 {
				p.executeFunc(c, files, p.hostStress[hostAddr], hostAddr)
				//getFiles = getFiles[:0]
				files = make(map[string]file)
			}

		case <-time.After(time.Second * 4):
			if len(files) == 0 {
				continue
			}
			p.executeFunc(c, files, p.hostStress[hostAddr], hostAddr)
			files = make(map[string]file)

		case <-p.ctx.Done():
			return
		}
	}
}
