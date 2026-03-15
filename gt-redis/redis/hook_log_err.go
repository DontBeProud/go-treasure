package gtredis

import (
	"context"
	"errors"

	gtlog "github.com/DontBeProud/go-treasure/gt-log"
	"github.com/redis/go-redis/v9"
)

var _ redis.Hook = (*hookLogErr)(nil)

type hookLogErr struct {
	logger gtlog.Logger
	name   string
	addr   string
}

func (h *hookLogErr) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *hookLogErr) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd Cmder) error {
		err := next(ctx, cmd)
		if err != nil && !errors.Is(err, redis.Nil) {
			h.logger.ErrorWContext(ctx, "msg", namespace+"_err", "name", h.name, "addr", h.addr, "detail", cmd.String())
		}
		return err
	}
}

func (h *hookLogErr) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []Cmder) error {
		err := next(ctx, cmds)
		for _, cmd := range cmds {
			if cmd.Err() != nil && !errors.Is(cmd.Err(), redis.Nil) {
				h.logger.ErrorWContext(ctx, "msg", namespace+"_err", "name", h.name, "addr", h.addr, "detail", cmd.String())
			}
		}
		return err
	}
}
