package socket

import (
	"EXSync/core/option/exsync/comm"
	"github.com/sirupsen/logrus"
)

// SendStat 构造错误反馈, 收到错误一方打印日志并立即终止当前命令的操作
func SendStat(s *Session, stat string) {
	command := comm.Command{
		Command: "",
		Type:    "",
		Method:  "",
		Data: map[string]any{
			"stat": stat,
		},
	}
	_, err := s.SendCommand(command, false, true)
	if err != nil {
		logrus.Errorf("SendStat: An error occurred while sending error message \"%s\"! %s", stat, err)
	}
	return
}
