package ioc

import (
	"github.com/redis/go-redis/v9"
	"webook/webook/config"
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
