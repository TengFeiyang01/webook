package ioc

import (
	"github.com/redis/go-redis/v9"
	"time"
	"webook/webook/config"
	"webook/webook/pkg/ratelimit"
)

var redisClient *redis.Client

func InitRedis() redis.Cmdable {
	if redisClient == nil {
		redisClient = redis.NewClient(&redis.Options{
			Addr: config.Config.Redis.Addr,
		})
	}
	return redisClient
}

func NewRateLimiter(interval time.Duration, rate int) ratelimit.Limiter {
	return ratelimit.NewRedisSlidingWindowLimiter(redisClient, interval, rate)
}
