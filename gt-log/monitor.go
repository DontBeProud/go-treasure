package gtlog

import (
	"context"
	"runtime"
	"time"
)

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
