//go:build wireinject

package main

import (
	artrepo "github.com/TengFeiyang01/webook/webook/article/repository"
	artcache "github.com/TengFeiyang01/webook/webook/article/repository/cache"
	artdao "github.com/TengFeiyang01/webook/webook/article/repository/dao"
	artsvc "github.com/TengFeiyang01/webook/webook/article/service"
	repository2 "github.com/TengFeiyang01/webook/webook/interactive/repository"
	cache2 "github.com/TengFeiyang01/webook/webook/interactive/repository/cache"
	dao2 "github.com/TengFeiyang01/webook/webook/interactive/repository/dao"
	service2 "github.com/TengFeiyang01/webook/webook/interactive/service"
	"github.com/TengFeiyang01/webook/webook/internal/repository"
	"github.com/TengFeiyang01/webook/webook/internal/repository/cache"
	"github.com/TengFeiyang01/webook/webook/internal/repository/dao"
	"github.com/TengFeiyang01/webook/webook/internal/service"
	"github.com/TengFeiyang01/webook/webook/internal/web"
	ijwt "github.com/TengFeiyang01/webook/webook/internal/web/jwt"
	"github.com/TengFeiyang01/webook/webook/ioc"
	"github.com/google/wire"
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
		//ioc.InitKafka,
		ioc.NewConsumers,
		//ioc.NewSyncProducer,
		ioc.InitRLockClient,

		//interactiveSvcSet,
		//ioc.InitIntrGRPCClient,

		// etcd
		ioc.InitEtcd,
		ioc.InitIntrGRPCClientV1,
		rankingServiceSet,
		ioc.InitJobs,
		ioc.InitRankingJob,

		//articleSvcSet,
		//ioc.InitArtGRPCClient,

		ioc.InitArtGRPCClientV1,

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
		//artevents.NewKafkaProducer,

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
