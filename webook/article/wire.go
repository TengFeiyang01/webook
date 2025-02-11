//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/webook/article/events"
	"webook/webook/article/grpc"
	"webook/webook/article/ioc"
	"webook/webook/article/repository"
	"webook/webook/article/repository/cache"
	"webook/webook/article/repository/dao"
	"webook/webook/article/service"
	usrdao "webook/webook/internal/repository/dao"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitKafka,
	ioc.InitRedis,
)

var articleSvcSet = wire.NewSet(
	dao.NewGORMArticleDAO,
	repository.NewCachedArticleRepository,
	service.NewArticleService,
	cache.NewArticleCache,
	usrdao.NewUserDAO,
)

func InitAPP() *App {
	wire.Build(
		thirdPartySet,
		articleSvcSet,
		grpc.NewArticleServiceServer,
		events.NewKafkaProducer,
		ioc.NewGRPCxServer,
		ioc.NewSyncProducer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
