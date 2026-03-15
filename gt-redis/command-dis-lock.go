package gtredis

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	redisV9 "github.com/redis/go-redis/v9"
)

// CommandDistributedLock 分布式锁指令集
type CommandDistributedLock interface {
	// AcquireDistributedLock 获取分布式锁
	AcquireDistributedLock(ctx context.Context, info *DistributedLockInfo) (DistributedLock, bool, error)
}

// DistributedLock 分布式锁抽象接口
type DistributedLock interface {
	// Renewal 续期
	Renewal(ctx context.Context) error
	// Release 释放
	Release(ctx context.Context) error
	// AutoRenewal 自动续期（异步调用）
	AutoRenewal(ctx context.Context, interval time.Duration, afterRenewal func(error), wg *sync.WaitGroup)
	// GetSuggestedAutoRenewalInterval 获取建议的自动续期间隔
	GetSuggestedAutoRenewalInterval() time.Duration
}

// DistributedLockInfo 分布式锁信息
type DistributedLockInfo struct {
	LockName   string        // lock key
	LockToken  string        // 锁令牌，即校验值
	LockExpire time.Duration // 锁有效期
}

// AcquireDistributedLock 获取分布式锁
func (c *client) AcquireDistributedLock(ctx context.Context, info *DistributedLockInfo) (DistributedLock, bool, error) {
	if info == nil {
		return nil, false, errors.New("分布式锁配置为空")
	}
	if info.LockName == "" {
		return nil, false, errors.New("分布式锁名为空")
	}
	if info.LockToken == "" {
		return nil, false, errors.New("分布式锁令牌/校验值为空")
	}
	if info.LockExpire <= 0 {
		return nil, false, errors.New("分布式锁过期时间无效")
	}
	res := c.GetClient(ctx).SetNX(ctx, info.LockName, info.LockToken, info.LockExpire)
	if res.Val() {
		return &distributedLock{
			r:        c,
			info:     info,
			quitOnce: sync.Once{},
		}, true, nil
	}
	return nil, false, res.Err()
}

var (
	scriptReleaseLock            = redisV9.NewScript(luaReleaseLock)
	scriptRenewalLock            = redisV9.NewScript(luaRenewalLock)
	errorLockTokenNotMatch       = errors.New("dis lock err: token does not match")
	errorLockReleaseProcedure    = errors.New("dis lock err: release lock fail. ")
	errorLockReleaseDeleteFail   = errors.New("dis lock err: delete key fail in the process of releasing the redis lock")
	errorLockRenewalProcedure    = errors.New("dis lock err: refresh lock fail")
	errorLockRenewalUnknownError = errors.New("dis lock err: unknown error occurred in the process of releasing the redis lock")
	errorLockHasBeenReleased     = errors.New("dis lock err: lock has been released")
)

type distributedLock struct {
	r        *client
	info     *DistributedLockInfo
	quitOnce sync.Once
	// atomic.Bool 保证读写原子性，避免 data race
	quit atomic.Bool
}

// Release 释放锁
func (l *distributedLock) Release(ctx context.Context) error {
	if l.quit.Load() {
		return nil
	}
	var code int64
	var err error
	l.quitOnce.Do(func() {
		code, err = scriptReleaseLock.Run(ctx, l.r, []string{l.info.LockName}, l.info.LockToken).Int64()
		if err != nil {
			err = errors.New(errorLockReleaseProcedure.Error() + err.Error())
			return
		}
		switch code {
		case -1:
			err = errorLockTokenNotMatch
		case -2:
			err = errorLockReleaseDeleteFail
		default:
			l.quit.Store(true)
		}
	})
	return err
}

// Renewal 续期
func (l *distributedLock) Renewal(ctx context.Context) error {
	if l.quit.Load() {
		return errorLockHasBeenReleased
	}
	code, err := scriptRenewalLock.Run(ctx, l.r, []string{l.info.LockName}, l.info.LockToken, l.info.LockExpire.Milliseconds()).Int64()
	if err != nil {
		return errorLockRenewalProcedure
	}
	switch code {
	case 0:
		return errorLockRenewalUnknownError
	case -1:
		return errorLockTokenNotMatch
	}
	return nil
}

// AutoRenewal 自动续期
func (l *distributedLock) AutoRenewal(ctx context.Context, interval time.Duration, afterRenewal func(error), wg *sync.WaitGroup) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if ddl, ok := ctx.Deadline(); ok && !ddl.After(time.Now()) {
				// 使用 return 而非 break，确保退出整个函数而非只退出 select
				return
			}
			if l.quit.Load() {
				return
			}
			err := l.Renewal(ctx)
			if afterRenewal != nil {
				afterRenewal(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// GetSuggestedAutoRenewalInterval 获取建议的自动续期间隔
func (l *distributedLock) GetSuggestedAutoRenewalInterval() time.Duration {
	return min(l.info.LockExpire/3, 3*time.Second)
}

var (
	luaReleaseLock = `
	-- return -1 if the token does not match
	if (redis.call('get', KEYS[1]) ~= ARGV[1]) then 
		return -1
	end

	-- release lock
	if (redis.call('del', KEYS[1]) ~= 1) then
		return -2
	end
	return 1
	`

	luaRenewalLock = `
	-- token不符合则返回失败
	if (redis.call('get', KEYS[1]) ~= ARGV[1]) then 
		return -1
	end
	return redis.call('pexpire', KEYS[1], ARGV[2])
	`
)
