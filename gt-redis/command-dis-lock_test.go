package gtredis

import (
	"context"
	"testing"
	"time"

	gtlog "github.com/DontBeProud/go-treasure/gt-log"
	gtconfpb "github.com/DontBeProud/go-treasure/pb/gt-conf-pb"
)

func TestRedis_AcquireDistributedLock(t *testing.T) {
	skipTest(t)

	ctx := context.Background()
	r, _, _ := NewClient(&gtconfpb.RedisConfig{
		Con: &gtconfpb.RedisConnectConfig{
			Addr:     "ip:port",
			Password: "your_password",
		},
	}, gtlog.GetLogger())

	l, success, err := r.AcquireDistributedLock(ctx, &DistributedLockInfo{
		LockName:   "test_dis_lock_name",
		LockToken:  "test_dis_lock_token",
		LockExpire: 15 * time.Second,
	})
	if err != nil {
		panic(err.Error())
	}
	if !success {
		gtlog.GetLogger().ErrorW("msg", "锁处于占用状态")
		return
	}

	go l.AutoRenewal(ctx, l.GetSuggestedAutoRenewalInterval(), func(err error) {
		if err != nil {
			gtlog.GetLogger().ErrorW("msg", "renewal err", "err", err.Error())
		} else {
			gtlog.GetLogger().Info("renewal success")
		}
	}, nil)

	time.Sleep(20 * time.Second)

	if err = l.Release(ctx); err != nil {
		gtlog.GetLogger().ErrorW("msg", "release err", "err", err.Error())
	} else {
		gtlog.GetLogger().Info("release success")
	}

	time.Sleep(10 * time.Second)
}
