package logger

import (
	"fmt"
	"os"
	"runtime"
)

const callerSkip = 1

func Fatalf(format string, args ...any) {
	if levelCheck(FATAL.level) {
		_, file, line, _ := runtime.Caller(callerSkip)
		logger.logFatal.Printf("%s:%d: %v", file, line, fmt.Sprintf(format, args...))
	}
	os.Exit(1)
}

func Errorf(format string, args ...any) {
	if levelCheck(ERROR.level) {
		_, file, line, _ := runtime.Caller(callerSkip)
		logger.logError.Printf("%s:%d: %v", file, line, fmt.Sprintf(format, args...))
	}
}

func Warnf(format string, args ...any) {
	if levelCheck(WARN.level) {
		_, file, line, _ := runtime.Caller(callerSkip)
		logger.logWarn.Printf("%s:%d: %v", file, line, fmt.Sprintf(format, args...))
	}
}

func Infof(format string, args ...any) {
	if levelCheck(INFO.level) {
		_, file, line, _ := runtime.Caller(callerSkip)
		logger.logInfo.Printf("%s:%d: %v", file, line, fmt.Sprintf(format, args...))
	}
}

func Debugf(format string, args ...any) {
	if levelCheck(DEBUG.level) {
		_, file, line, _ := runtime.Caller(callerSkip)
		logger.logDebug.Printf("%s:%d: %v", file, line, fmt.Sprintf(format, args...))
	}
}

func Fatal(args ...any) {
	if levelCheck(FATAL.level) {
		_, file, line, _ := runtime.Caller(callerSkip)
		logger.logFatal.Printf("%s:%d: %v", file, line, args)
	}
	os.Exit(1)
}

func Error(args ...any) {
	if levelCheck(ERROR.level) {
		_, file, line, _ := runtime.Caller(callerSkip)
		logger.logError.Printf("%s:%d: %v", file, line, args)
	}
}

func Warn(args ...any) {
	if levelCheck(WARN.level) {
		_, file, line, _ := runtime.Caller(callerSkip)
		logger.logWarn.Printf("%s:%d: %v", file, line, args)
	}
}

func Info(args ...any) {
	if levelCheck(INFO.level) {
		_, file, line, _ := runtime.Caller(callerSkip)
		logger.logInfo.Printf("%s:%d: %v", file, line, args)
	}
}

func Debug(args ...any) {
	if levelCheck(DEBUG.level) {
		_, file, line, _ := runtime.Caller(callerSkip)
		logger.logDebug.Printf("%s:%d: %v", file, line, args)
	}
}
