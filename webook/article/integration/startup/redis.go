package startup

import (
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}
