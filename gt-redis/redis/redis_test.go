package gtredis

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	gtlog "github.com/DontBeProud/go-treasure/gt-log"
	"github.com/redis/go-redis/v9"
)

func skipTest(t *testing.T) {
	t.Helper()
	if os.Getenv("RUN_TEST") != "true" {
		t.Skip("skipping test; set RUN_TEST=true to run")
	}
}

// TestRedis 测试 Redis
func TestRedis(t *testing.T) {
	skipTest(t)
	ctx := context.Background()

	r, _, _ := NewRedis(
		Logger(gtlog.GetLogger()),
		Addr("127.0.0.1:6379"),
		Username(""),
		Password(""),
		DialTimeout(5*time.Second),
		ReadTimeout(3*time.Second),
		WriteTimeout(3*time.Second),
		PoolSize(10*runtime.NumCPU()),
		MinIdleConns(0),
		MaxConnAge(0),
		PoolTimeout(4*time.Second),
		IdleTimeout(5*time.Minute),
		Name("test"),
		LogErr(true),
		LogSlow(1*time.Millisecond),
	)

	// 单次执行命令
	r.Do(ctx, "set", "hello", "world")
	r.Do(ctx, "get", "hello")

	// 错误命令
	r.Do(ctx, "set2", "hello")

	// 批量执行命令
	_, _ = r.Pipelined(ctx, func(pipeliner Pipeliner) error {
		pipeliner.Do(ctx, "set", "hello1", "world1")
		pipeliner.Do(ctx, "set", "hello2", "world2")
		pipeliner.Do(ctx, "set", "hello3", "world3")
		pipeliner.Do(ctx, "get", "hello1")
		pipeliner.Do(ctx, "get", "hello2")
		pipeliner.Do(ctx, "get", "hello3")
		return nil
	})

	// 执行 lua 脚本示例
	keys := []string{"my_counter"}
	values := []interface{}{+1}
	incrBy.Load(ctx, r)
	incrBy.Run(ctx, r, keys, values...)

	time.Sleep(5 * time.Second)
}

// TestSentinel 测试 Sentinel
func TestSentinel(t *testing.T) {
	skipTest(t)
	ctx := context.Background()

	r, _, _ := NewRedis(
		Logger(gtlog.GetLogger()),
		ClusterAddr([]string{"192.168.182.121:26379", "192.168.182.121:36379", "192.168.182.121:46379"}),
		MasterName("master"),
		Password("your_password"),
		DialTimeout(5*time.Second),
		ReadTimeout(3*time.Second),
		WriteTimeout(3*time.Second),
		PoolSize(10*runtime.NumCPU()),
		MinIdleConns(0),
		MaxConnAge(0),
		PoolTimeout(4*time.Second),
		IdleTimeout(5*time.Minute),
		Name("test"),
		LogErr(true),
		LogSlow(1*time.Millisecond),
	)

	r.Do(ctx, "set", "hello", "world")

	time.Sleep(5 * time.Second)
}

var incrBy = redis.NewScript(`
local key = KEYS[1]
local change = ARGV[1]
local value = redis.call("GET", key)
if not value then
  value = 0
end
value = value + change
redis.call("SET", key, value)
return value
`)
