package socket

import (
	"net"
)

// IsCloseConnect 判断当前错误是否为向一个已关闭的socket发送数据
func IsCloseConnect(err error) bool {
	netErr, ok := err.(*net.OpError)
	if ok && netErr.Err.Error() == "use of closed network connection" {
		return true
	}
	return false
}
