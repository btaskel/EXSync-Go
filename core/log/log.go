package loger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Log = NewLog()

func NewLog() *logrus.Logger {
	log := logrus.New()
	file, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Errorf("Failed to log to file, using default stderr")
		os.Exit(1)
	}
	// 设置日志格式
	log.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
	return log
}
