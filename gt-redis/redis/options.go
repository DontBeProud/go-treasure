package gtredis

import (
	"strings"
	"time"

	gtlog "github.com/DontBeProud/go-treasure/gt-log"
	"github.com/redis/go-redis/v9"
)

type Option func(r *Redis)

// Logger set logger
func Logger(logger gtlog.Logger) Option {
	return func(r *Redis) {
		r.logger = logger
	}
}

// Addr set addr (host:port, default 127.0.0.1:6379)
func Addr(addr string) Option {
	return func(r *Redis) {
		r.addr = addr
		r.clusterAddr = []string{r.addr}
	}
}

// ClusterAddr set cluster addresses
func ClusterAddr(addrs []string) Option {
	return func(r *Redis) {
		addr := strings.Builder{}
		for _, a := range addrs {
			_, _ = addr.WriteString("[")
			_, _ = addr.WriteString(a)
			_, _ = addr.WriteString("]")
		}
		r.addr = addr.String()
		// go-redis 判断集群的方式是 len(addrs) > 1，所以这里 append 一个 addrs
		r.clusterAddr = append(addrs, addrs...)
	}
}

// MasterName set masterName (sentinel mode)
func MasterName(masterName string) Option {
	return func(r *Redis) {
		r.masterName = masterName
	}
}

// Username set username
func Username(username string) Option {
	return func(r *Redis) {
		r.username = username
	}
}

// Password set password
func Password(password string) Option {
	return func(r *Redis) {
		r.password = password
	}
}

// SentinelUsername set sentinelUsername
func SentinelUsername(sentinelUsername string) Option {
	return func(r *Redis) {
		r.sentinelUsername = sentinelUsername
	}
}

// SentinelPassword set sentinelPassword
func SentinelPassword(sentinelPassword string) Option {
	return func(r *Redis) {
		r.sentinelPassword = sentinelPassword
	}
}

// DialTimeout set dialTimeout (default 5s)
func DialTimeout(dialTimeout time.Duration) Option {
	return func(r *Redis) {
		r.dialTimeout = dialTimeout
	}
}

// ReadTimeout set readTimeout (default 3s)
func ReadTimeout(readTimeout time.Duration) Option {
	return func(r *Redis) {
		r.readTimeout = readTimeout
	}
}

// WriteTimeout set writeTimeout (default equals readTimeout)
func WriteTimeout(writeTimeout time.Duration) Option {
	return func(r *Redis) {
		r.writeTimeout = writeTimeout
	}
}

// PoolSize set poolSize (default 10 * NumCPU)
func PoolSize(poolSize int) Option {
	return func(r *Redis) {
		r.poolSize = poolSize
	}
}

// MinIdleConns set minIdleConns (default 0)
func MinIdleConns(minIdleConns int) Option {
	return func(r *Redis) {
		r.minIdleConns = minIdleConns
	}
}

// MaxConnAge set maxConnAge (default 0, no expiry)
func MaxConnAge(maxConnAge time.Duration) Option {
	return func(r *Redis) {
		r.maxConnAge = maxConnAge
	}
}

// PoolTimeout set poolTimeout (default readTimeout + 1s)
func PoolTimeout(poolTimeout time.Duration) Option {
	return func(r *Redis) {
		r.poolTimeout = poolTimeout
	}
}

// IdleTimeout set idleTimeout (default 5min, -1 to disable)
func IdleTimeout(idleTimeout time.Duration) Option {
	return func(r *Redis) {
		r.idleTimeout = idleTimeout
	}
}

// DB set db index
func DB(db int) Option {
	return func(r *Redis) {
		r.db = db
	}
}

// ReadOnly set readonly (commands route to replicas)
func ReadOnly(readOnly bool) Option {
	return func(r *Redis) {
		r.readOnly = readOnly
	}
}

// RouteRandomly set routeRandomly
func RouteRandomly(routeRandomly bool) Option {
	return func(r *Redis) {
		r.routeRandomly = routeRandomly
	}
}

// RouteByLatency set routeByLatency
func RouteByLatency(routeByLatency bool) Option {
	return func(r *Redis) {
		r.routeByLatency = routeByLatency
	}
}

// Name set instance name (used to distinguish instances in logs/metrics)
func Name(name string) Option {
	return func(r *Redis) {
		r.name = name
	}
}

// LogErr enable/disable error logging (default true)
func LogErr(logErr bool) Option {
	return func(r *Redis) {
		r.logErr = logErr
	}
}

// LogSlow set slow-query logging threshold (default 250ms, 0 to disable)
func LogSlow(logSlow time.Duration) Option {
	return func(r *Redis) {
		r.logSlow = logSlow
	}
}

// Hooks set external hooks for unified interception
func Hooks(hooks ...redis.Hook) Option {
	return func(r *Redis) {
		r.hooks = hooks
	}
}
