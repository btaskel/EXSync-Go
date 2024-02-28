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

//type HostQueue struct {
//	Queue map[Host]clientComm.GetFile
//}
//
//type Pool struct {
//	queue   chan map[Host]clientComm.GetFile
//	dynamic chan []clientComm.GetFile
//	clients chan []*client.Client
//}
//

//
//func parseIndex(*base.IndexFile) {
//
//}

//// Add 向池里添加新同步空间队列
//func (p *Pool) Add(fileQueue *HostQueue) error {
//
//}
