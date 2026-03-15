package gtredis

import (
	"context"
	"runtime"
	"sync"
	"time"

	gtlog "github.com/DontBeProud/go-treasure/gt-log"
	"github.com/redis/go-redis/v9"
)

const namespace = "redis_client"

var (
	_ gtRedis  = (*Redis)(nil)
	_ Scripter = (*Redis)(nil)
)

type gtRedis interface {
	Do(ctx context.Context, args ...interface{}) *Cmd
	Pipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error)
	TxPipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error)
	Subscribe(ctx context.Context, channels ...string) *PubSub
	PSubscribe(ctx context.Context, channels ...string) *PubSub
}

// Redis wraps go-redis UniversalClient with logging hooks.
type Redis struct {
	client redis.UniversalClient
	logger gtlog.Logger

	// Connection config
	addr             string
	clusterAddr      []string
	masterName       string
	username         string
	password         string
	sentinelUsername string
	sentinelPassword string
	db               int
	readOnly         bool
	routeRandomly    bool
	routeByLatency   bool

	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration

	// Pool config
	poolSize     int
	minIdleConns int
	maxConnAge   time.Duration
	poolTimeout  time.Duration
	idleTimeout  time.Duration

	// Logging
	name    string
	logErr  bool
	logSlow time.Duration

	closed     chan struct{}
	onceClosed sync.Once
	hooks      []redis.Hook
}

var (
	addrDefault           = "127.0.0.1:6379"
	readOnlyDefault       = true
	routeRandomlyDefault  = false
	routeByLatencyDefault = false
	dialTimeoutDefault    = 5 * time.Second
	readTimeoutDefault    = 3 * time.Second
	poolSizeDefault       = 10 * runtime.NumCPU()
	idleTimeoutDefault    = 5 * time.Minute
	logErrDefault         = true
	logSlowDefault        = 250 * time.Millisecond
)

// NewRedis creates a Redis instance. All operations go through the wrapped Redis object.
// r: wrapped Redis object; cleanup: call on graceful shutdown; err: creation error.
func NewRedis(options ...Option) (r *Redis, cleanup func(), err error) {
	r = &Redis{
		logger: gtlog.GetLogger(),

		addr:           addrDefault,
		readOnly:       readOnlyDefault,
		routeRandomly:  routeRandomlyDefault,
		routeByLatency: routeByLatencyDefault,
		dialTimeout:    dialTimeoutDefault,
		readTimeout:    readTimeoutDefault,

		poolSize:    poolSizeDefault,
		idleTimeout: idleTimeoutDefault,

		logErr:  logErrDefault,
		logSlow: logSlowDefault,

		closed: make(chan struct{}, 1),
	}
	r.writeTimeout = r.readTimeout
	r.poolTimeout = r.readTimeout + time.Second

	for _, option := range options {
		option(r)
	}

	cleanup = func() {
		r.onceClosed.Do(func() {
			close(r.closed)
			if r != nil && r.client != nil {
				_ = r.client.Close()
			}
		})
	}

	r.client = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:            r.clusterAddr,
		MasterName:       r.masterName,
		Username:         r.username,
		Password:         r.password,
		SentinelUsername: r.sentinelUsername,
		SentinelPassword: r.sentinelPassword,
		DialTimeout:      r.dialTimeout,
		ReadTimeout:      r.readTimeout,
		WriteTimeout:     r.writeTimeout,
		DB:               r.db,
		ReadOnly:         r.readOnly,
		RouteRandomly:    r.routeRandomly,
		RouteByLatency:   r.routeByLatency,

		PoolSize:        r.poolSize,
		MinIdleConns:    r.minIdleConns,
		ConnMaxLifetime: r.maxConnAge,
		PoolTimeout:     r.poolTimeout,
		ConnMaxIdleTime: r.idleTimeout,
	})

	for _, hook := range r.hooks {
		r.client.AddHook(hook)
	}

	if r.logErr {
		r.client.AddHook(&hookLogErr{logger: r.logger, name: r.name, addr: r.addr})
	}

	if r.logSlow > 0 {
		r.client.AddHook(&hookLogSlow{logger: r.logger, logSlow: r.logSlow, name: r.name, addr: r.addr})
	}

	err = r.client.Ping(context.Background()).Err()
	if err != nil {
		return
	}

	return
}

// GetClient returns the underlying UniversalClient.
func (r *Redis) GetClient(_ context.Context) redis.UniversalClient {
	return r.client
}

// Do executes a Redis command.
func (r *Redis) Do(ctx context.Context, args ...interface{}) *Cmd {
	return r.client.Do(ctx, args...)
}

// Pipelined executes commands in a pipeline.
func (r *Redis) Pipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	return r.client.Pipelined(ctx, fn)
}

// TxPipelined executes commands in a transactional pipeline.
func (r *Redis) TxPipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	return r.client.TxPipelined(ctx, fn)
}

// Subscribe subscribes to channels.
func (r *Redis) Subscribe(ctx context.Context, channels ...string) *PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// PSubscribe subscribes to channels matching patterns.
func (r *Redis) PSubscribe(ctx context.Context, channels ...string) *PubSub {
	return r.client.PSubscribe(ctx, channels...)
}

// Eval implements redis.Scripter.
func (r *Redis) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd {
	return r.client.Eval(ctx, script, keys, args...)
}

// EvalSha implements redis.Scripter.
func (r *Redis) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd {
	return r.client.EvalSha(ctx, sha1, keys, args...)
}

// ScriptExists implements redis.Scripter.
func (r *Redis) ScriptExists(ctx context.Context, hashes ...string) *BoolSliceCmd {
	return r.client.ScriptExists(ctx, hashes...)
}

// ScriptLoad implements redis.Scripter.
func (r *Redis) ScriptLoad(ctx context.Context, script string) *StringCmd {
	return r.client.ScriptLoad(ctx, script)
}

// EvalRO implements redis.Scripter.
func (r *Redis) EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	return r.client.EvalRO(ctx, script, keys, args...)
}

// EvalShaRO implements redis.Scripter.
func (r *Redis) EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	return r.client.EvalShaRO(ctx, sha1, keys, args...)
}
