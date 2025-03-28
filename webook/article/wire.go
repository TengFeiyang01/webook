//go:build wireinject

package main

import (
	"github.com/TengFeiyang01/webook/webook/article/events"
	"github.com/TengFeiyang01/webook/webook/article/grpc"
	"github.com/TengFeiyang01/webook/webook/article/ioc"
	"github.com/TengFeiyang01/webook/webook/article/repository"
	"github.com/TengFeiyang01/webook/webook/article/repository/cache"
	"github.com/TengFeiyang01/webook/webook/article/repository/dao"
	"github.com/TengFeiyang01/webook/webook/article/service"
	usrdao "github.com/TengFeiyang01/webook/webook/internal/repository/dao"
	"github.com/google/wire"
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
