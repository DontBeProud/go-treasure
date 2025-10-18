package gtlog

import (
	"context"
	"testing"
	"time"

	"github.com/bytedance/sonic"
)

// TestLogger 测试 loggerObj
func TestLogger(t *testing.T) { //nolint
	logger, _, _ := NewLogger(
		// 日志是否写控制台（默认为 false）
		LogToStdout(true),
		// 日志是否写文件（默认为 true）
		LogToFile(false),
		// 日志文件目录（默认为 loggerObj）
		Path("logger"),
		// 日志文件名（默认为 log.log）
		FileName("log.log"),
		// 日志滚动时间（默认为 1 小时）
		RotationTime(time.Hour),
		// 日志保存时间（默认为 180 天）
		FileAge(time.Hour*24*180),
		// 日志最低输出等级（默认为 debug）
		Level("debug"),
		// 固定输出的 key-value
		With("key", "value"),
		// 是否输出 JSON 格式（默认为 true）
		FormatJSON(false),
	)

	ctx := context.Background()

	// debug
	logger.Debug("debug")
	logger.DebugW("msg", "debug", "k1", "v1")
	logger.DebugContext(ctx, "debug")
	logger.DebugWContext(ctx, "msg", "debug", "k1", "v1")

	// info
	logger.Info("info")
	logger.InfoW("msg", "info", "k1", "v1")
	logger.InfoContext(ctx, "info")
	logger.InfoWContext(ctx, "msg", "info", "k1", "v1")

	// warn
	logger.Warn("warn")
	logger.WarnW("msg", "warn", "k1", "v1")
	logger.WarnContext(ctx, "warn")
	logger.WarnWContext(ctx, "msg", "warn", "k1", "v1")

	// error
	logger.Error("error")
	logger.ErrorW("msg", "error", "k1", "v1")
	logger.ErrorContext(ctx, "error")
	logger.ErrorWContext(ctx, "msg", "error", "k1", "v1")

	// 增加多种数据结构的输出单元测试
	// 基本数据
	i := 1
	s := "s"
	b := true
	// map
	m := []string{"m1", "m2"}
	// struct
	type User struct {
		Name string
		Age  uint32
	}
	u := &User{
		Name: "name",
		Age:  18,
	}
	// json
	j, _ := sonic.ConfigStd.Marshal(u)
	logger.InfoW("i", i, "s", s, "b", b, "m", m, "u", u, "j", string(j))
	logger.InfoWContext(ctx, "i", i, "s", s, "b", b, "m", m, "u", u, "j", string(j))

	// impl kratos logger
	_ = logger.LogContext(ctx, LevelInfo, "kratos k1", "kratos v1")
}

// TestLoggerWithJSONFormat 测试 JSON 格式 loggerObj
func TestLoggerWithJSONFormat(t *testing.T) { //nolint
	logger, _, _ := NewLogger(
		// 日志是否写控制台（默认为 false）
		LogToStdout(true),
		// 日志是否写文件（默认为 true）
		LogToFile(false),
		// 日志文件目录（默认为 loggerObj）
		Path("logger"),
		// 日志文件名（默认为 log.log）
		FileName("log.log"),
		// 日志滚动时间（默认为 1 小时）
		RotationTime(time.Hour),
		// 日志保存时间（默认为 180 天）
		FileAge(time.Hour*24*180),
		// 日志最低输出等级（默认为 debug）
		Level("debug"),
		// 固定输出的 key-value
		With("key", "value"),
		// 是否输出 JSON 格式（默认为 true）
		FormatJSON(true),
	)

	ctx := context.Background()
	ctx = SetLogContext(ctx, LogContext{
		"trace_id": "trace_id_123456",
		"span_id":  "span_id_123456",
	})

	// debug
	logger.Debug("debug")
	logger.DebugW("msg", "debug", "k1", "v1")
	logger.DebugContext(ctx, "debug")
	logger.DebugWContext(ctx, "msg", "debug", "k1", "v1")

	// info
	logger.Info("info")
	logger.InfoW("msg", "info", "k1", "v1")
	logger.InfoContext(ctx, "info")
	logger.InfoWContext(ctx, "msg", "info", "k1", "v1")

	// warn
	logger.Warn("warn")
	logger.WarnW("msg", "warn", "k1", "v1")
	logger.WarnContext(ctx, "warn")
	logger.WarnWContext(ctx, "msg", "warn", "k1", "v1")

	// error
	logger.Error("error")
	logger.ErrorW("msg", "error", "k1", "v1")
	logger.ErrorContext(ctx, "error")
	logger.ErrorWContext(ctx, "msg", "error", "k1", "v1")

	// 增加多种数据结构的输出单元测试
	// 基本数据
	i := 1
	s := "s"
	b := true
	// map
	m := []string{"m1", "m2"}
	// struct
	type User struct {
		Name string
		Age  uint32
	}
	u := &User{
		Name: "name",
		Age:  18,
	}
	// json
	j, _ := sonic.ConfigStd.Marshal(u)
	logger.InfoW("i", i, "s", s, "b", b, "m", m, "u", u, "j", string(j))
	logger.InfoWContext(ctx, "i", i, "s", s, "b", b, "m", m, "u", u, "j", string(j))

	// impl kratos logger
	_ = logger.LogContext(ctx, LevelInfo, "kratos k1", "kratos v1")
}
