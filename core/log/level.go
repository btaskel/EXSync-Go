package loger

import (
	"github.com/sirupsen/logrus"
	"strings"
)

func FormatLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		Log.Level = logrus.DebugLevel
	case "info":
		Log.Level = logrus.InfoLevel
	case "warning":
		Log.Level = logrus.WarnLevel
	case "error":
		Log.Level = logrus.ErrorLevel
	case "fatal":
		Log.Level = logrus.FatalLevel
	case "panic":
		Log.Level = logrus.PanicLevel
	default:
		Log.Level = logrus.InfoLevel
	}
}
