package gtmysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	gtlog "github.com/DontBeProud/go-treasure/gt-log"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// LoggerConfig gorm日志配置
type LoggerConfig struct {
	LogLevel                      logger.LogLevel // 日志级别
	SlowThreshold                 time.Duration   // 慢SQL阈值
	DontIgnoreRecordNotFoundError bool            // 不忽略 ErrRecordNotFound 错误
	DontIgnoreKeyDuplicateError   bool            // 不忽略唯一键冲突错误
	FilterParams                  bool            // 打印SQL时隐藏参数
}

// NewGormLogger 新建标准化gorm日志对象
func NewGormLogger(loggerCore gtlog.Logger, cfg *LoggerConfig) (logger.Interface, error) {
	if loggerCore == nil {
		return nil, errors.New("invalid loggerCore")
	}
	if cfg == nil {
		return nil, errors.New("invalid LoggerConfig")
	}
	return &gormLogger{
		LoggerConfig: *cfg,
		loggerCore:   loggerCore,
	}, nil
}

type gormLogger struct {
	LoggerConfig
	loggerCore gtlog.Logger
}

// LogMode log mode
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info print info
func (l *gormLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.loggerCore.Info(fmt.Sprintf(msg, append([]interface{}{utils.FileWithLineNum()}, data...)...))
	}
}

// Warn print warn messages
func (l *gormLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.loggerCore.Warn(fmt.Sprintf(msg, append([]interface{}{utils.FileWithLineNum()}, data...)...))
	}
}

// Error print error messages
func (l *gormLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.loggerCore.Error(fmt.Sprintf(msg, append([]interface{}{utils.FileWithLineNum()}, data...)...))
	}
}

// Trace print sql trace
func (l *gormLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	params := []interface{}{
		"msg", "mysql trace",
		"invoker_path", utils.FileWithLineNum(),
		"sql", sql,
		"latency", float64(elapsed.Nanoseconds()) / 1e6,
	}
	if rows != -1 {
		params = append(params, "rows", rows)
	} else {
		params = append(params, "rows", "-")
	}

	if err != nil {
		if l.LogLevel >= logger.Error &&
			(!errors.Is(err, logger.ErrRecordNotFound) || l.DontIgnoreRecordNotFoundError) &&
			(!IsErrorDuplicateKey(err) || l.DontIgnoreKeyDuplicateError) {
			params = append(params, "error", err)
			l.loggerCore.ErrorW(params...)
		}
		return
	}

	if elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn {
		params = append(params, "slow_sql", true)
		l.loggerCore.WarnW(params...)
	} else if l.LogLevel >= logger.Info {
		l.loggerCore.InfoW(params...)
	}
}

// IsErrorDuplicateKey 判断是否是唯一键冲突错误
func IsErrorDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	_err := &mysqlDriver.MySQLError{}
	if !errors.As(err, &_err) {
		return false
	}
	return fmt.Sprintf("%s", _err.SQLState) == "23000"
}
