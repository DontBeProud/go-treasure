package gtredis

import (
	"context"
	"fmt"
	"strings"
	"time"

	gtlog "github.com/DontBeProud/go-treasure/gt-log"
	"github.com/redis/go-redis/v9"
)

var _ redis.Hook = (*hookLogSlow)(nil)

type hookLogSlow struct {
	logger  gtlog.Logger
	logSlow time.Duration
	name    string
	addr    string
}

type ctxKey string

var startCtxKey ctxKey = namespace + ":start"

func (h *hookLogSlow) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *hookLogSlow) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if ctx.Value(startCtxKey) == nil {
			ctx = context.WithValue(ctx, startCtxKey, time.Now())
		}
		err := next(ctx, cmd)
		ts := time.Since(ctx.Value(startCtxKey).(time.Time))
		if ts > h.logSlow {
			h.logger.WarnWContext(ctx, "msg", namespace+"_slow", "name", h.name, "addr", h.addr, "ts", fmt.Sprintf("%dms", ts.Milliseconds()), "detail", cmd.String())
		}
		return err
	}
}

func (h *hookLogSlow) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		if ctx.Value(startCtxKey) == nil {
			ctx = context.WithValue(ctx, startCtxKey, time.Now())
		}
		err := next(ctx, cmds)
		ts := time.Since(ctx.Value(startCtxKey).(time.Time))
		if ts > h.logSlow {
			cmdName := strings.Builder{}
			for _, cmd := range cmds {
				_, _ = cmdName.WriteString("[")
				_, _ = cmdName.WriteString(cmd.Name())
				_, _ = cmdName.WriteString("]")
			}
			h.logger.WarnWContext(ctx, "msg", namespace+"_slow", "name", h.name, "addr", h.addr, "ts", fmt.Sprintf("%dms", ts.Milliseconds()), "detail", cmdName.String())
		}
		return err
	}
}
