package manage

import (
	serverOption "EXSync/core/option/exsync/trans"
)

type TaskManage struct {
	Trans map[string]serverOption.TranTask
}

type Lock struct {
	FileLock map[string]struct{}
}
