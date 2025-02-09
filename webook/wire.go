//go:build wireinject

package main

import (
	"github.com/google/wire"
	events2 "webook/webook/interactive/events"
	repository2 "webook/webook/interactive/repository"
	cache2 "webook/webook/interactive/repository/cache"
	dao2 "webook/webook/interactive/repository/dao"
	service2 "webook/webook/interactive/service"
	events "webook/webook/internal/events/article"
	"webook/webook/internal/repository"
	artrepo "webook/webook/internal/repository/article"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/repository/dao/article"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	ijwt "webook/webook/internal/web/jwt"
	"webook/webook/ioc"
)

var interactiveSvcSet = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

var rankingServiceSet = wire.NewSet(
	repository.NewCachedRankingRepository,
	cache.NewRankingRedisCache,
	service.NewBatchRankingService,
)

func InitApp() *App {
	wire.Build(
		// 初始化 DB
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewConsumers,
		ioc.NewSyncProducer,
		ioc.InitRLockClient,

		rankingServiceSet,
		ioc.InitJobs,
		ioc.InitRankingJob,
		ioc.InitIntrGRPCClient,
		interactiveSvcSet,

		// 初始化 DAO
		dao.NewUserDAO,

		// 初始化 cache
		cache.NewRedisUserCache,
		//cache.NewLocalCodeCache,
		cache.NewRedisCodeCache,
		cache.NewArticleCache,
		cache.NewRankingLocalCache,

		// 初始化 repository
		repository.NewUserRepository,
		repository.NewCodeRepository,
		artrepo.NewCachedArticleRepository,

		// consumer
		events2.NewInteractiveReadEventBatchConsumer,
		events.NewKafkaProducer,

		// 初始化 service
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,

		ioc.InitSMSService,
		ioc.InitOAuth2WechatService,
		ioc.NewWechatHandlerConfig,

		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		article.NewGORMArticleDAO,
		ijwt.NewRedisJWT,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
