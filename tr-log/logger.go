package trlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	kLog "github.com/go-kratos/kratos/v2/log"
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	normalLogger = iota
	shadowLogger
	testLogger
)

var (
	_ Logger      = (*loggerObj)(nil)
	_ kLog.Logger = (*loggerObj)(nil)
)

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	DebugW(kv ...interface{})
	InfoW(kv ...interface{})
	WarnW(kv ...interface{})
	ErrorW(kv ...interface{})

	DebugContext(ctx context.Context, msg string)
	InfoContext(ctx context.Context, msg string)
	WarnContext(ctx context.Context, msg string)
	ErrorContext(ctx context.Context, msg string)
	DebugWContext(ctx context.Context, kv ...interface{})
	InfoWContext(ctx context.Context, kv ...interface{})
	WarnWContext(ctx context.Context, kv ...interface{})
	ErrorWContext(ctx context.Context, kv ...interface{})
	LogContext(ctx context.Context, level Lvl, kvs ...interface{}) error
}

type loggerObj struct {
	// Wrapper
	logger       *zap.Logger // 正常日志组件
	loggerShadow *zap.Logger // 影子日志组件
	loggerTest   *zap.Logger // 测试日志组件

	// Config
	setGlobal    bool          // 是否设置为全局 loggerObj，仅设置一次（默认为 true）
	logToStdout  bool          // 日志是否写控制台（默认为 false）
	logToFile    bool          // 日志是否写文件（默认为 true）
	path         string        // 日志文件目录（默认为 Logger）
	fileName     string        // 日志文件名（默认为 log.log）
	rotationTime time.Duration // 日志滚动时间（默认为 1 小时）
	fileAge      time.Duration // 日志保存时间（默认为 180 天）
	level        string        // 日志最低输出等级（默认为 debug）
	kv           []interface{} // 固定输出的 key-value
	formatJSON   bool          // 是否输出 JSON 格式（默认为 true）
	callSkip     int           // 调用者层级

	// Filter
	filterKey   map[interface{}]struct{}
	filterValue map[interface{}]struct{}
	filter      func(l LogLevel, kvs ...interface{}) bool

	// Buffered Writer
	bufferedSize          int           // 缓冲区大小，默认为 16k（与 slog 保持相对一致），设置为 0 则为同步阻塞输出日志，否则为异步输出，缓冲区写满时输出一次
	bufferedFlushInterval time.Duration // Flush 缓冲区时间间隔，默认为 10s
}

var (
	// 配置的默认值
	setGlobal             = true
	logToStdout           = false
	logToFile             = true
	path                  = "logger"
	fileName              = "log.log"
	rotationTime          = time.Hour
	fileAge               = time.Hour * 24 * 180
	level                 = "debug"
	formatJSON            = true
	callSkip              = 2
	bufferedSize          = 16 * 1024
	bufferedFlushInterval = time.Second * 1
)

// NewLogger 创建 loggerObj 操作对象，对外只暴露这一个方法，所有操作都通过封装过的 loggerObj 对象来处理
// logger 封装过的 loggerObj 操作对象；err 创建对象时的异常
func NewLogger(options ...Option) (ll Logger, cleanup func(), err error) {
	l := &loggerObj{
		setGlobal:             true,
		logToStdout:           false,
		logToFile:             true,
		path:                  path,
		fileName:              fileName,
		rotationTime:          rotationTime,
		fileAge:               fileAge,
		level:                 level,
		filterKey:             make(map[interface{}]struct{}),
		filterValue:           make(map[interface{}]struct{}),
		formatJSON:            true,
		callSkip:              callSkip,
		bufferedSize:          bufferedSize,
		bufferedFlushInterval: bufferedFlushInterval,
	}

	for _, option := range options {
		option(l)
	}

	var encoder zapcore.Encoder
	encoder = zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "ts",
		CallerKey:      "file",
		SkipLineEnding: false,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(t.Format("2006-01-02 15:04:05.000 -0700"))
			encoder.AppendString("0")
		},
		EncodeCaller: func(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(caller.TrimmedPath())
			encoder.AppendString("-")
		},
		ConsoleSeparator: " ",
	})

	// 滚动日志
	linkName := filepath.Join(l.path, l.fileName)
	name := l.fileName
	suffix := ""
	i := strings.LastIndex(l.fileName, ".")
	if i >= 0 {
		name = l.fileName[:i]
		suffix = l.fileName[i:]
	}
	fileName := filepath.Join(l.path, fmt.Sprintf("%s-%s%s", name, "%Y-%m-%d-%H", suffix))
	rotate, errRotateLogs := rotateLogs.New(
		fileName,
		rotateLogs.WithLinkName(linkName),
		rotateLogs.WithMaxAge(l.fileAge),
		rotateLogs.WithRotationTime(l.rotationTime),
	)
	if errRotateLogs != nil {
		return nil, nil, errRotateLogs
	}

	// 创建 zap loggerObj
	level, errParseLevel := zapcore.ParseLevel(l.level)
	if errParseLevel != nil {
		return nil, nil, errParseLevel
	}
	var cores []zapcore.Core
	if l.logToStdout {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
	}
	if l.logToFile {
		if l.bufferedSize > 0 {
			bws := zapcore.BufferedWriteSyncer{
				WS:            zapcore.AddSync(rotate),
				Size:          l.bufferedSize,
				FlushInterval: l.bufferedFlushInterval,
			}
			cores = append(cores, zapcore.NewCore(encoder, &bws, level))
		} else {
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(rotate), level))
		}
	}
	core := zapcore.NewTee(cores...)
	zapLog := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(l.callSkip))
	l.logger = zapLog

	cleanup = func() {
		if l != nil && l.logger != nil {
			_ = l.logger.Sync()
		}
		if l != nil && l.loggerShadow != nil {
			_ = l.loggerShadow.Sync()
		}
		if l != nil && l.loggerTest != nil {
			_ = l.loggerTest.Sync()
		}
	}

	if l.setGlobal {
		once.Do(func() {
			globalLogger = l
		})
	}

	return l, cleanup, err
}

func (l *loggerObj) Path() string {
	return l.path
}

func (l *loggerObj) FileName() string {
	return l.fileName
}

// defaultMessageKey default message key.
var defaultMessageKey = "msg"

// Debug 输出 Debug 日志
func (l *loggerObj) Debug(msg string) {
	l.logf(context.TODO(), zapcore.DebugLevel, defaultMessageKey, msg)
}

// Info 输出 Info 日志
func (l *loggerObj) Info(msg string) {
	l.logf(context.TODO(), zapcore.InfoLevel, defaultMessageKey, msg)
}

// Warn 输出 Warn 日志
func (l *loggerObj) Warn(msg string) {
	l.logf(context.TODO(), zapcore.WarnLevel, defaultMessageKey, msg)
}

// Error 输出 Error 日志
func (l *loggerObj) Error(msg string) {
	l.logf(context.TODO(), zapcore.ErrorLevel, defaultMessageKey, msg)
}

// DebugW 按照 kv 输出 Debug 日志
func (l *loggerObj) DebugW(kv ...interface{}) {
	l.logf(context.TODO(), zapcore.DebugLevel, kv...)
}

// InfoW 按照 kv 输出 Info 日志
func (l *loggerObj) InfoW(kv ...interface{}) {
	l.logf(context.TODO(), zapcore.InfoLevel, kv...)
}

// WarnW 按照 kv 输出 Warn 日志
func (l *loggerObj) WarnW(kv ...interface{}) {
	l.logf(context.TODO(), zapcore.WarnLevel, kv...)
}

// ErrorW 按照 kv 输出 Error 日志
func (l *loggerObj) ErrorW(kv ...interface{}) {
	l.logf(context.TODO(), zapcore.ErrorLevel, kv...)
}

// DebugContext 输出 Debug 日志
func (l *loggerObj) DebugContext(ctx context.Context, msg string) {
	l.logf(ctx, zapcore.DebugLevel, mergeFields(ctx, defaultMessageKey, msg)...)
}

// InfoContext 输出 Info 日志
func (l *loggerObj) InfoContext(ctx context.Context, msg string) {
	l.logf(ctx, zapcore.InfoLevel, mergeFields(ctx, defaultMessageKey, msg)...)
}

// WarnContext 输出 Warn 日志
func (l *loggerObj) WarnContext(ctx context.Context, msg string) {
	l.logf(ctx, zapcore.WarnLevel, mergeFields(ctx, defaultMessageKey, msg)...)
}

// ErrorContext 输出 Error 日志
func (l *loggerObj) ErrorContext(ctx context.Context, msg string) {
	l.logf(ctx, zapcore.ErrorLevel, mergeFields(ctx, defaultMessageKey, msg)...)
}

// DebugWContext 按照 kv 输出 Debug 日志
func (l *loggerObj) DebugWContext(ctx context.Context, kv ...interface{}) {
	l.logf(ctx, zapcore.DebugLevel, mergeFields(ctx, kv...)...)
}

// InfoWContext 按照 kv 输出 Info 日志
func (l *loggerObj) InfoWContext(ctx context.Context, kv ...interface{}) {
	l.logf(ctx, zapcore.InfoLevel, mergeFields(ctx, kv...)...)
}

// WarnWContext 按照 kv 输出 Warn 日志
func (l *loggerObj) WarnWContext(ctx context.Context, kv ...interface{}) {
	l.logf(ctx, zapcore.WarnLevel, mergeFields(ctx, kv...)...)
}

// ErrorWContext 按照 kv 输出 Error 日志
func (l *loggerObj) ErrorWContext(ctx context.Context, kv ...interface{}) {
	l.logf(ctx, zapcore.ErrorLevel, mergeFields(ctx, kv...)...)
}

// Log 实现 kratos 的日志接口
func (l *loggerObj) Log(level kLog.Level, kvs ...interface{}) error {
	return l.LogContext(context.TODO(), level, kvs...)
}

func (l *loggerObj) LogContext(ctx context.Context, level Lvl, kvs ...interface{}) error {
	switch level {
	case LevelDebug:
		l.logf(ctx, zapcore.DebugLevel, mergeFields(ctx, kvs)...)
	case LevelInfo:
		l.logf(ctx, zapcore.InfoLevel, mergeFields(ctx, kvs)...)
	case LevelWarn:
		l.logf(ctx, zapcore.WarnLevel, mergeFields(ctx, kvs)...)
	case LevelError:
		l.logf(ctx, zapcore.ErrorLevel, mergeFields(ctx, kvs)...)
	case LevelFatal:
		l.logf(ctx, zapcore.WarnLevel, "msg", "level unavailable", "level", level)
	}
	return nil
}

// logf 格式化输出日志
func (l *loggerObj) logf(ctx context.Context, level zapcore.Level, args ...interface{}) { //nolint
	// 日志级别小于配置时不处理日志格式化，提前返回，提高性能
	lvl, _ := zapcore.ParseLevel(l.level)
	if !lvl.Enabled(level) {
		return
	}

	// 参数数量不是 2 的倍数时打印 warn 日志
	if len(args)%2 != 0 {
		l.logger.Warn(fmt.Sprint("msg=args must appear in pairs", args))
		return
	}

	fkv := make([]interface{}, 0, len(l.kv)+len(args))
	fkv = append(fkv, l.kv...)
	fkv = append(fkv, args...)

	if l.filter != nil && l.filter(level, fkv...) {
		return
	}

	var text []byte
	var fields []zapcore.Field
	if l.formatJSON {
		buf := make(map[string]interface{})
		for i := 1; i < len(fkv); i += 2 {
			// 兼容 kratos 的 log.Valuer 接口
			k := fkv[i-1]
			v := fkv[i]
			if vv, ok := fkv[i].(kLog.Valuer); ok {
				v = vv(ctx)
				// v 有值才打印
				if v == nil || v == "" {
					continue
				}
			}
			if _, ok := k.(string); ok {
				if _, ok := l.filterKey[k]; ok {
					v = fuzzyStr
				}
			}
			if _, ok := v.(string); ok {
				if _, ok := l.filterValue[v]; ok {
					v = fuzzyStr
				}
			}

			var (
				vv  string
				err error
			)
			// 如果是基本数据类型则直接输出，不转 string（搭配 kibana 来检索）
			t := reflect.TypeOf(v)
			if t != nil && (t.Kind() >= reflect.Bool && t.Kind() <= reflect.Complex128 || t.Kind() == reflect.String) {
				buf[cast.ToString(k)] = v
			} else if v == nil || (t != nil && t.Kind() == reflect.Interface && reflect.ValueOf(v).IsNil()) {
				buf[cast.ToString(k)] = "null"
			} else if vv, err = cast.ToStringE(v); err != nil {
				// 如果是可以转 JSON 的结构体，则尝试转 JSON 字符串
				vj, _ := sonic.ConfigStd.Marshal(v)
				buf[cast.ToString(k)] = json.RawMessage(vj)
			} else {
				buf[cast.ToString(k)] = vv
			}
		}
		text, _ = sonic.ConfigStd.Marshal(buf)
	} else {
		var buf bytes.Buffer
		for i := 1; i < len(fkv); i += 2 {
			// 兼容 kratos 的 log.Valuer 接口
			k := fkv[i-1]
			v := fkv[i]
			if vv, ok := fkv[i].(kLog.Valuer); ok {
				v = vv(ctx)
			}
			if _, ok := k.(string); ok {
				if _, ok := l.filterKey[k]; ok {
					v = fuzzyStr
				}
			}
			if vv, ok := v.(string); ok {
				if _, ok := l.filterValue[v]; ok {
					v = fuzzyStr
				}
				// len(vv) 大于 0 才打印
				if len(vv) == 0 {
					continue
				}
			}
			if vv, ok := v.(json.RawMessage); ok {
				v = string(vv)
				if _, ok := l.filterValue[v]; ok {
					v = fuzzyStr
				}
				// len(vv) 大于 0 才打印
				if len(vv) == 0 {
					continue
				}
			}
			_, _ = fmt.Fprintf(&buf, "%s=%v&", k, v)
		}

		text = buf.Bytes()
		// 截断最后的 &
		if len(text) > 0 {
			text = text[0 : len(text)-1]
		}
	}

	switch level {
	case zapcore.DebugLevel:
		l.logger.Debug(string(text), fields...)
	case zapcore.InfoLevel:
		l.logger.Info(string(text), fields...)
	case zapcore.WarnLevel:
		l.logger.Warn(string(text), fields...)
	case zapcore.ErrorLevel:
		l.logger.Error(string(text), fields...)
	default:
	}
}
