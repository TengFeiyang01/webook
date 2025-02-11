//go:build wireinject

package main

import (
	"github.com/google/wire"
	artevents "webook/webook/article/events"
	artrepo "webook/webook/article/repository"
	artcache "webook/webook/article/repository/cache"
	artdao "webook/webook/article/repository/dao"
	artsvc "webook/webook/article/service"
	events2 "webook/webook/interactive/events"
	repository2 "webook/webook/interactive/repository"
	cache2 "webook/webook/interactive/repository/cache"
	dao2 "webook/webook/interactive/repository/dao"
	service2 "webook/webook/interactive/service"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
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

var articleSvcSet = wire.NewSet(
	artcache.NewArticleCache,
	artrepo.NewCachedArticleRepository,
	artsvc.NewArticleService,
	artdao.NewGORMArticleDAO,
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

		articleSvcSet,
		ioc.InitArtGRPCClient,

		// 初始化 DAO
		dao.NewUserDAO,

		// 初始化 cache
		cache.NewRedisUserCache,
		//cache.NewLocalCodeCache,
		cache.NewRedisCodeCache,

		// 初始化 repository
		repository.NewUserRepository,
		repository.NewCodeRepository,

		// consumer
		events2.NewInteractiveReadEventBatchConsumer,
		artevents.NewKafkaProducer,

		// 初始化 service
		service.NewUserService,
		service.NewCodeService,

		ioc.InitSMSService,
		ioc.InitOAuth2WechatService,
		ioc.NewWechatHandlerConfig,

		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		ijwt.NewRedisJWT,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
