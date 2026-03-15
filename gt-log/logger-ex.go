package gtlog

// LoggerExOption is an option for LoggerEx.
type LoggerExOption func(l *LoggerEx)

// LoggerEx 增强版日志对象，支持 SupportContext 开关
type LoggerEx struct {
	Logger
	SupportContext bool
}

// GetLoggerEx 获取全局 Logger 的增强版包装
func GetLoggerEx(options ...LoggerExOption) *LoggerEx {
	return NewLoggerEx(GetLogger(), options...)
}

// NewLoggerEx 创建增强版日志对象
func NewLoggerEx(logger Logger, options ...LoggerExOption) *LoggerEx {
	l := &LoggerEx{
		Logger:         logger,
		SupportContext: true,
	}
	for _, option := range options {
		option(l)
	}
	return l
}

// WithSupportContextDisabled 禁用 context 传递
func WithSupportContextDisabled(l *LoggerEx) {
	l.SupportContext = false
}
