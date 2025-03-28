package ioc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"github.com/TengFeiyang01/webook/webook/internal/repository"
	"github.com/TengFeiyang01/webook/webook/internal/repository/cache"
	"github.com/TengFeiyang01/webook/webook/internal/service"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"github.com/TengFeiyang01/webook/webook/pkg/redisx"
)

func InitUserHandler(repo repository.UserRepository) service.UserService {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return service.NewUserService(repo, logger.NewZapLogger(l))
}

// InitUserCache 配合 PrometheusHook 使用
func InitUserCache(client *redis.Client) cache.UserCache {
	client.AddHook(redisx.NewPrometheusHook(
		prometheus.SummaryOpts{
			Namespace: "ytf",
			Subsystem: "webook",
			Name:      "user",
			Help:      "统计 Redis 的接口",
			ConstLabels: map[string]string{
				"biz": "user",
			},
		}))
	panic("do not call this !!!")
}
