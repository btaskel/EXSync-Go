package logger

import (
	"log"
	"os"
)

// NewLog 在release时, 设置为 NULL 将会被编译器优化掉, 这将不会影响性能
var logger = NewLog(DEBUG)

var (
	NULL  = Level{level: 10, color: White, prefix: ""}
	DEBUG = Level{level: 1, color: Blue, prefix: "DEBUG-"}
	INFO  = Level{level: 2, color: Green, prefix: "INFO-"}
	WARN  = Level{level: 3, color: Yellow, prefix: "WARN-"}
	ERROR = Level{level: 4, color: Red, prefix: "ERROR-"}
	FATAL = Level{level: 5, color: DeepRed, prefix: "FATAL-"}
)

type Level struct {
	level  int
	color  string
	prefix string
}

type Logger struct {
	level    int
	logFatal *log.Logger
	logError *log.Logger
	logWarn  *log.Logger
	logInfo  *log.Logger
	logDebug *log.Logger
}

func NewLog(level Level) *Logger {
	_logger := new(Logger)

	if NULL.level > level.level && level.level > FATAL.level {
		panic("Logging level not supported")
	}

	_logger.level = level.level
	_logger.logFatal = log.New(os.Stdout, FATAL.color+FATAL.prefix, log.Ltime)
	_logger.logError = log.New(os.Stdout, ERROR.color+ERROR.prefix, log.Ltime)
	_logger.logWarn = log.New(os.Stdout, WARN.color+WARN.prefix, log.Ltime)
	_logger.logInfo = log.New(os.Stdout, INFO.color+INFO.prefix, log.Ltime)
	_logger.logDebug = log.New(os.Stdout, DEBUG.color+DEBUG.prefix, log.Ltime)
	return _logger
}

func levelCheck(level int) bool {
	if logger == nil {
		panic("The Logging is not initialized")
	}

	if logger.level <= level {
		return true
	} else {
		return false
	}
}
