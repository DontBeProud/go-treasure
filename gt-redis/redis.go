package gtredis

import (
	"errors"

	gtlog "github.com/DontBeProud/go-treasure/gt-log"
	gtredis "github.com/DontBeProud/go-treasure/gt-redis/redis"
	gtconfpb "github.com/DontBeProud/go-treasure/pb/gt-conf-pb"
)

// Client 基于平台redis库的二次封装，便于后续的扩展方法
type Client interface {
	// CommandUniversal 通用指令集
	CommandUniversal
	// CommandDistributedLock 分布式锁指令集
	CommandDistributedLock
}

type client struct {
	*gtredis.Redis
	logger gtlog.Logger
}

// FromRedis 基于平台redis对象生成封装redis对象
func FromRedis(r *gtredis.Redis, logger gtlog.Logger) Client {
	return &client{r, logger}
}

// NewClient 创建redis对象
func NewClient(config *gtconfpb.RedisConfig, logger gtlog.Logger) (r Client, cleanup func(), err error) {
	return newClient(config, logger)
}

func newClient(config *gtconfpb.RedisConfig, logger gtlog.Logger) (r Client, cleanup func(), err error) {
	if config == nil {
		return nil, func() {}, errors.New("redis config is nil")
	}

	opts := []gtredis.Option{
		gtredis.MasterName(config.MasterName),
		gtredis.DB(int(config.Db)),
		gtredis.ReadOnly(!config.DisableReadOnly),
		gtredis.RouteRandomly(config.RouteRandomly),
		gtredis.RouteByLatency(config.RouteByLatency),
		gtredis.MinIdleConns(int(config.MinIdleCons)),
	}

	if logger != nil {
		opts = append(opts, gtredis.Logger(logger))
	}

	if con := config.Con; con != nil {
		if con.Addr != "" {
			opts = append(opts, gtredis.Addr(con.Addr))
		}
		if len(con.ClusterAddr) > 0 {
			opts = append(opts, gtredis.ClusterAddr(con.ClusterAddr))
		}
		opts = append(opts, gtredis.Username(con.UserName))
		opts = append(opts, gtredis.Password(con.Password))
		opts = append(opts, gtredis.SentinelUsername(con.SentinelUserName))
		opts = append(opts, gtredis.SentinelPassword(con.SentinelPassword))
	}

	if config.PoolSize != nil {
		opts = append(opts, gtredis.PoolSize(int(config.GetPoolSize())))
	}
	if config.DialTimeout != nil {
		opts = append(opts, gtredis.DialTimeout(config.DialTimeout.AsDuration()))
	}
	if config.ReadTimeout != nil {
		opts = append(opts, gtredis.ReadTimeout(config.ReadTimeout.AsDuration()))
	}
	if config.WriteTimeout != nil {
		opts = append(opts, gtredis.WriteTimeout(config.WriteTimeout.AsDuration()))
	}
	if config.MaxConAge != nil {
		opts = append(opts, gtredis.MaxConnAge(config.MaxConAge.AsDuration()))
	}
	if config.PoolTimeout != nil {
		opts = append(opts, gtredis.PoolTimeout(config.PoolTimeout.AsDuration()))
	}
	if config.IdleTimeout != nil {
		opts = append(opts, gtredis.IdleTimeout(config.IdleTimeout.AsDuration()))
	}

	obj, cleanup, err := gtredis.NewRedis(opts...)
	if err != nil {
		return nil, func() {}, err
	}
	return &client{obj, logger}, cleanup, nil
}
