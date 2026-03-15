package gtredis
import (
"context"
"os"
"testing"
"time"
gtlog "github.com/DontBeProud/go-treasure/gt-log"
gtconfpb "github.com/DontBeProud/go-treasure/pb/gt-conf-pb"
redisV9 "github.com/redis/go-redis/v9"
)
func skipTest(t *testing.T) {
t.Helper()
if os.Getenv("RUN_TEST") != "true" {
t.Skip("skipping test; set RUN_TEST=true to run")
}
}
func TestNewRedis(t *testing.T) {
skipTest(t)
ctx := context.Background()
r, _, _ := NewClient(&gtconfpb.RedisConfig{
Con: &gtconfpb.RedisConnectConfig{
Addr:     "ip:port",
Password: "your_password",
},
}, gtlog.GetLogger())
r.Do(ctx, "set", "hello", "world")
r.Do(ctx, "get", "hello")
_, _ = r.Pipelined(ctx, func(pipeliner Pipeliner) error {
pipeliner.Do(ctx, "set", "hello1", "world1")
pipeliner.Do(ctx, "set", "hello2", "world2")
pipeliner.Do(ctx, "get", "hello1")
pipeliner.Do(ctx, "get", "hello2")
return nil
})
keys := []string{"my_counter"}
values := []interface{}{+1}
incrBy := redisV9.NewScript(` + "`" + `
local key = KEYS[1]
local change = ARGV[1]
local value = redis.call("GET", key)
if not value then
  value = 0
end
value = value + change
redis.call("SET", key, value)
return value
` + "`" + `)
incrBy.Load(ctx, r)
incrBy.Run(ctx, r, keys, values...)
time.Sleep(5 * time.Second)
}
