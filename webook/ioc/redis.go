package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"time"
	"github.com/TengFeiyang01/webook/webook/pkg/ratelimit"
)

var redisClient *redis.Client

func InitRedis() redis.Cmdable {
	//addr := viper.GetString("redis.addr")
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("redis", &cfg)
	if err != nil {
		panic(err)
	}
	if redisClient == nil {
		redisClient = redis.NewClient(&redis.Options{
			Addr: cfg.Addr,
		})
	}
	return redisClient
}

func InitRLockClient(cmd redis.Cmdable) *rlock.Client {
	return rlock.NewClient(cmd)
}

func NewRateLimiter(interval time.Duration, rate int) ratelimit.Limiter {
	return ratelimit.NewRedisSlidingWindowLimiter(redisClient, interval, rate)
}
