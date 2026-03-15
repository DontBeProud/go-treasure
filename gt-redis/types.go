package gtredis

import gtredis "github.com/DontBeProud/go-treasure/gt-redis/redis"

// Re-export common types from the redis subpackage for convenience.
type (
	Pipeliner = gtredis.Pipeliner
	Cmder     = gtredis.Cmder
	PubSub    = gtredis.PubSub
)
