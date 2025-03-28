//go:build wireinject

package main

import (
	"github.com/TengFeiyang01/webook/webook/interactive/events"
	"github.com/TengFeiyang01/webook/webook/interactive/grpc"
	"github.com/TengFeiyang01/webook/webook/interactive/ioc"
	"github.com/TengFeiyang01/webook/webook/interactive/repository"
	"github.com/TengFeiyang01/webook/webook/interactive/repository/cache"
	"github.com/TengFeiyang01/webook/webook/interactive/repository/dao"
	"github.com/TengFeiyang01/webook/webook/interactive/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitKafka,
	ioc.InitRedis,
)

var interactiveSvcSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	service.NewInteractiveService,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
)

func InitAPP() *App {
	wire.Build(interactiveSvcSet,
		thirdPartySet,
		grpc.NewInteractiveServiceServer,
		events.NewInteractiveEventConsumer,
		ioc.NewConsumers,
		ioc.NewGRPCxServer,
		wire.Struct(new(App), "*"))
	return new(App)
}
