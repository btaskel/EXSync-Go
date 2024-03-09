package trans

import "context"

type TranTask struct {
	DstAddr  string              // 目标主机.
	Progress *float32            // 百分比进度.
	Cancel   *context.CancelFunc // 取消任务.
}

type TranBat *struct {
}
