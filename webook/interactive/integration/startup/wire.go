//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"github.com/TengFeiyang01/webook/webook/interactive/grpc"
	"github.com/TengFeiyang01/webook/webook/interactive/repository"
	"github.com/TengFeiyang01/webook/webook/interactive/repository/cache"
	"github.com/TengFeiyang01/webook/webook/interactive/repository/dao"
	"github.com/TengFeiyang01/webook/webook/interactive/service"
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

func InitInteractiveGRPCServer() *grpc.InteractiveServiceServer {
	wire.Build(thirdPartySet, interactiveSvcSet, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}
