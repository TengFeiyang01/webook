//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/webook/interactive/events"
	"webook/webook/interactive/grpc"
	"webook/webook/interactive/ioc"
	"webook/webook/interactive/repository"
	"webook/webook/interactive/repository/cache"
	"webook/webook/interactive/repository/dao"
	"webook/webook/interactive/service"
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
