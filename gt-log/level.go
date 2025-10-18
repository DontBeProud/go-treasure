package gtlog

import (
	kratosLog "github.com/go-kratos/kratos/v2/log"
)

type Lvl = kratosLog.Level

const (
	// LevelDebug is logger debug level.
	LevelDebug = kratosLog.LevelDebug
	// LevelInfo is logger info level.
	LevelInfo = kratosLog.LevelInfo
	// LevelWarn is logger warn level.
	LevelWarn = kratosLog.LevelWarn
	// LevelError is logger error level.
	LevelError = kratosLog.LevelError
	// LevelFatal is logger fatal level
	LevelFatal = kratosLog.LevelFatal
)
