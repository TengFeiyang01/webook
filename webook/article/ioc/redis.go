package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
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
