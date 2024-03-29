package socket

import (
	loger "EXSync/core/log"
	"EXSync/core/option/exsync/comm"
)

// SendStat 构造错误反馈, 收到错误一方打印日志并立即终止当前命令的操作
// 无论执行结果如何, 该函数均不会抛出错误, 但是上层必须在该函数执行后立即返回
func SendStat(s *Session, stat string) {
	command := comm.Command{
		Data: map[string]any{
			"stat": stat,
		},
	}
	_, err := s.SendCommand(command, false, true)
	if err != nil {
		loger.Log.Errorf("SendStat: An error occurred while sending error message \"%s\"! %s", stat, err)
	}
	return
}
