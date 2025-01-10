//go:build wireinject

package main

import (
	"github.com/google/wire"
	events "webook/webook/internal/events/article"
	"webook/webook/internal/repository"
	artrepo "webook/webook/internal/repository/article"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/cache/user"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/repository/dao/article"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	ijwt "webook/webook/internal/web/jwt"
	"webook/webook/ioc"
)

func InitApp() *App {
	wire.Build(
		// 初始化 DB
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewConsumers,
		ioc.NewSyncProducer,

		// 初始化 DAO
		dao.NewUserDAO,
		dao.NewGORMInteractiveDAO,

		// 初始化 cache
		user.NewRedisUserCache,
		//cache.NewLocalCodeCache,
		cache.NewRedisCodeCache,
		cache.NewArticleCache,
		cache.NewInteractiveRedisCache,

		// 初始化 repository
		repository.NewUserRepository,
		repository.NewCodeRepository,
		artrepo.NewCachedArticleRepository,
		repository.NewCachedInteractiveRepository,

		// consumer
		events.NewInteractiveReadEventBatchConsumer,
		events.NewKafkaProducer,

		// 初始化 service
		service.NewUserService,
		service.NewCodeService,
		service.NewInteractiveService,
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
