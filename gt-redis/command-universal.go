package gtredis

import (
	"context"

	gtredis "github.com/DontBeProud/go-treasure/gt-redis/redis"
	redisV9 "github.com/redis/go-redis/v9"
)

// CommandUniversal 通用指令集
type CommandUniversal interface {
	// GetClient 获取redis客户端对象
	GetClient(ctx context.Context) redisV9.UniversalClient
	// Do 执行 Redis 命令
	Do(ctx context.Context, args ...interface{}) *gtredis.Cmd
	// Pipelined 批量执行 Redis 命令
	Pipelined(ctx context.Context, fn func(gtredis.Pipeliner) error) ([]gtredis.Cmder, error)
	// TxPipelined 事务批量执行 Redis 命令
	TxPipelined(ctx context.Context, fn func(gtredis.Pipeliner) error) ([]gtredis.Cmder, error)
	// Subscribe Redis 订阅
	Subscribe(ctx context.Context, channels ...string) *gtredis.PubSub
	// PSubscribe Redis 订阅
	PSubscribe(ctx context.Context, channels ...string) *gtredis.PubSub
	// Eval 实现 redis.Scripter 接口
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *gtredis.Cmd
	// EvalSha 实现 redis.Scripter 接口
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *gtredis.Cmd
	// ScriptExists 实现 redis.Scripter 接口
	ScriptExists(ctx context.Context, hashes ...string) *gtredis.BoolSliceCmd
	// ScriptLoad 实现 redis.Scripter 接口
	ScriptLoad(ctx context.Context, script string) *gtredis.StringCmd
	// EvalRO 实现 redis.Scripter 接口
	EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *gtredis.Cmd
	// EvalShaRO 实现 redis.Scripter 接口
	EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *gtredis.Cmd
}
