package startup

import (
	"github.com/TengFeiyang01/webook/webook/internal/service/sms"
	"github.com/TengFeiyang01/webook/webook/internal/service/sms/memory"
	"github.com/TengFeiyang01/webook/webook/internal/service/sms/ratelimit"
	"github.com/TengFeiyang01/webook/webook/internal/service/sms/retryable"
	limiter "github.com/TengFeiyang01/webook/webook/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitSMSService(cmd redis.Cmdable) sms.Service {
	svc := ratelimit.NewRatelimitSMSService(memory.NewService(),
		limiter.NewRedisSlidingWindowLimiter(cmd, time.Second, 100))
	return retryable.NewService(svc, 3)
}
