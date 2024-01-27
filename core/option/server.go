package option

import "EXSync/core/internal/exsync/client"

type ConnectManage struct {
	ID         string
	ClientMark string
	Client     client.Client
}
