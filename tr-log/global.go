package trlog

import (
	"sync"
)

var (
	globalLogger Logger
	once         sync.Once
)

func init() {
	globalLogger, _, _ = NewLogger(
		// 是否设置为全局 loggerObj，仅设置一次（默认为 true）
		SetGlobal(false),
		// 日志是否写控制台（默认为 false）
		LogToStdout(true),
		// 日志是否写文件（默认为 true）
		LogToFile(false),
	)
}

// GetLogger returns global logger appliance as logger in current process.
func GetLogger() Logger {
	return globalLogger
}
