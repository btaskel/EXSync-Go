package logger

import "testing"

func TestPrint(t *testing.T) {
	Debug("测试")
	Info("测试")
	Warn("测试")
	Error("测试")
	Fatal("测试")
}

func TestFormatPrint(t *testing.T) {
	Debugf("测试: %v", 123)
	Infof("测试: %v", 123)
	Warnf("测试: %v", 123)
	Errorf("测试: %v", 123)
	Fatalf("测试: %v", 123)
}
