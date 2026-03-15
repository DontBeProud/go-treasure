package gtlog

import "context"

// SetLogField 将需要写入到日志的KV添加到context
func SetLogField(ctx context.Context, key string, value interface{}) context.Context {
	return SetLogContext(ctx, LogContext{key: value})
}

type LogContext map[string]interface{}

// SetLogContext 将需要写入到日志的KV添加到context
func SetLogContext(ctx context.Context, fields LogContext) context.Context {
	f := GetLogContext(ctx)
	for k, v := range fields {
		(*f)[k] = v
	}
	ctx = context.WithValue(ctx, LogFields, f)
	return ctx
}

func GetLogContext(ctx context.Context) *LogContext {
	fields, ok := ctx.Value(LogFields).(*LogContext)
	if !ok {
		f := make(LogContext)
		fields = &f
	}
	return fields
}

func (l *LoggerEx) DebugContext(ctx context.Context, msg string) {
	l.Logger.DebugWContext(ctx, mergeFields(ctx, defaultMessageKey, msg)...)
}

func (l *LoggerEx) InfoContext(ctx context.Context, msg string) {
	l.Logger.InfoWContext(ctx, mergeFields(ctx, defaultMessageKey, msg)...)
}

func (l *LoggerEx) WarnContext(ctx context.Context, msg string) {
	l.Logger.WarnWContext(ctx, mergeFields(ctx, defaultMessageKey, msg)...)
}

func (l *LoggerEx) ErrorContext(ctx context.Context, msg string) {
	l.Logger.ErrorWContext(ctx, mergeFields(ctx, defaultMessageKey, msg)...)
}

func (l *LoggerEx) DebugwContext(ctx context.Context, kv ...interface{}) {
	l.Logger.DebugWContext(ctx, mergeFields(ctx, kv...)...)
}

func (l *LoggerEx) InfowContext(ctx context.Context, kv ...interface{}) {
	l.Logger.InfoWContext(ctx, mergeFields(ctx, kv...)...)
}

func (l *LoggerEx) WarnwContext(ctx context.Context, kv ...interface{}) {
	l.Logger.WarnWContext(ctx, mergeFields(ctx, kv...)...)
}

func (l *LoggerEx) ErrorwContext(ctx context.Context, kv ...interface{}) {
	l.Logger.ErrorWContext(ctx, mergeFields(ctx, kv...)...)
}

func (l *LoggerEx) LogContext(ctx context.Context, level Lvl, kvs ...interface{}) error {
	return l.Logger.LogContext(ctx, level, mergeFields(ctx, kvs)...)
}

type key string

const LogFields key = "GS_LOG_FIELDS"

func mergeFields(ctx context.Context, kvs ...interface{}) []interface{} {
	ctxFields := GetLogContext(ctx)
	ctxLen := len(*ctxFields)
	fields := make([]interface{}, ctxLen*2)
	idx := 0
	for k, v := range *ctxFields {
		fields[idx+0] = k
		fields[idx+1] = v
		idx += 2
	}
	fields = append(fields, kvs...)
	return fields
}
