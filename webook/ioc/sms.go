package ioc

import (
	"github.com/redis/go-redis/v9"
	"time"
	"webook/webook/internal/service/sms"
	"webook/webook/internal/service/sms/memory"
	"webook/webook/internal/service/sms/ratelimit"
	limiter "webook/webook/pkg/ratelimit"
)

func InitSMSService(cmd redis.Cmdable) sms.Service {
	svc := ratelimit.NewRatelimitSMSService(memory.NewService(),
		limiter.NewRedisSlidingWindowLimiter(cmd, time.Second, 100))
	return svc
}
