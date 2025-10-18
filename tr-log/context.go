package trlog

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

type key string

const LogFields key = "TR_LOG_FIELDS"

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
