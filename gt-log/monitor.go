package gtlog

import (
	"context"
	"runtime"
	"runtime/debug"
	"time"
)

// MonitorLogger 程序状态监控日志对象
type MonitorLogger interface {
	// ExecuteMonitorTask 执行监控任务
	// interval: 轮询间隔
	// callback: 自定义回调函数
	ExecuteMonitorTask(ctx context.Context, interval time.Duration, callback func(*runtime.MemStats, Logger))
	// GetLogger 获取底层日志对象
	GetLogger() Logger
}

// NewMonitorLogger 创建标准化监控日志对象
func NewMonitorLogger(options ...Option) (MonitorLogger, func(), error) {
	logObj, cleanup, err := NewLogger(options...)
	if err != nil {
		return nil, func() {}, err
	}
	return &monitorLogger{Logger: logObj}, cleanup, nil
}

// monitorLogger 标准化监控日志对象
type monitorLogger struct {
	Logger
}

// GetLogger 获取底层日志对象
func (l *monitorLogger) GetLogger() Logger {
	return l.Logger
}

// ExecuteMonitorTask 执行监控任务
func (l *monitorLogger) ExecuteMonitorTask(ctx context.Context, interval time.Duration, callback func(*runtime.MemStats, Logger)) {
	defer func() {
		if err := recover(); err != nil {
			l.ErrorW("monitor err", err, "stack", string(debug.Stack()))
		}
	}()

	RunMonitor(ctx, l.Logger, interval, callback)
}

// RunMonitor 运行监控，直到 ctx 取消
func RunMonitor(ctx context.Context, l Logger, interval time.Duration, callback func(*runtime.MemStats, Logger)) {
	memStatus := runtime.MemStats{}
	tick := time.NewTicker(interval)
	for {
		runtime.ReadMemStats(&memStatus)

		if callback != nil {
			callback(&memStatus, l)
		} else {
			l.InfoW("current goroutine", runtime.NumGoroutine(),
				"heap", float64(memStatus.HeapInuse)/1024.0/1024.0,
				"stack", float64(memStatus.StackInuse)/1024.0/1024.0,
			)
		}

		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			continue
		}
	}
}
