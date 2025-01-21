//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"webook/webook/interactive/repository"
	"webook/webook/interactive/repository/cache"
	"webook/webook/interactive/repository/dao"
	"webook/webook/interactive/service"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB, InitLogger,
	NewSyncProducer,
	InitKafka,
)

var interactiveSvcSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	service.NewInteractiveService,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
)

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdPartySet, interactiveSvcSet)
	return service.NewInteractiveService(nil)
}
