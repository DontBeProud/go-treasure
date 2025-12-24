package gtlog

import (
	"time"
)

type (
	Option func(l *loggerObj)
)

// LogToStdout set logToStdout
func LogToStdout(logToStdout bool) Option {
	return func(l *loggerObj) {
		l.logToStdout = logToStdout
	}
}

// LogToFile set logToFile
func LogToFile(logToFile bool) Option {
	return func(l *loggerObj) {
		l.logToFile = logToFile
	}
}

// Path set path
func Path(path string) Option {
	return func(l *loggerObj) {
		l.path = path
	}
}

// FileName set fileName
func FileName(fileName string) Option {
	return func(l *loggerObj) {
		l.fileName = fileName
	}
}

// RotationTime set rotationTime
func RotationTime(rotationTime time.Duration) Option {
	return func(l *loggerObj) {
		l.rotationTime = rotationTime
	}
}

// FileAge set fileAge
func FileAge(fileAge time.Duration) Option {
	return func(l *loggerObj) {
		l.fileAge = fileAge
	}
}

// Level set level
func Level(level string) Option {
	return func(l *loggerObj) {
		l.level = level
	}
}

// With set with kv
func With(kv ...interface{}) Option {
	return func(l *loggerObj) {
		l.kv = append(l.kv, kv...)
	}
}

// SetGlobal set setGlobal
func SetGlobal(setGlobal bool) Option {
	return func(l *loggerObj) {
		l.setGlobal = setGlobal
	}
}

// FormatJSON set formatJSON
func FormatJSON(formatJSON bool) Option {
	return func(l *loggerObj) {
		l.formatJSON = formatJSON
	}
}

func BeautyJson(beautyJson bool) Option {
	return func(l *loggerObj) {
		l.beautyJSON = beautyJson
	}
}

// CallSkip set callSkip
func CallSkip(callSkip int) Option {
	return func(l *loggerObj) {
		l.callSkip = callSkip
	}
}

// FilterKey with filter key.
func FilterKey(k ...string) Option {
	return func(l *loggerObj) {
		for _, v := range k {
			l.filterKey[v] = struct{}{}
		}
	}
}

// FilterValue with filter value.
func FilterValue(v ...string) Option {
	return func(l *loggerObj) {
		for _, v := range v {
			l.filterValue[v] = struct{}{}
		}
	}
}

// FilterFunc with filter func.
func FilterFunc(f func(level LogLevel, kvs ...interface{}) bool) Option {
	return func(l *loggerObj) {
		l.filter = f
	}
}

// BufferedSize set bufferedSize
func BufferedSize(bufferedSize int) Option {
	return func(l *loggerObj) {
		l.bufferedSize = bufferedSize
	}
}

// BufferedFlushInterval set bufferedFlushInterval
func BufferedFlushInterval(bufferedFlushInterval time.Duration) Option {
	return func(l *loggerObj) {
		l.bufferedFlushInterval = bufferedFlushInterval
	}
}
