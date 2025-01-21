//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	article2 "webook/webook/internal/events/article"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/article"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	artdao "webook/webook/internal/repository/dao/article"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	ijwt "webook/webook/internal/web/jwt"
	"webook/webook/ioc"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	ioc.InitRedis, InitDB, InitLogger,
	ioc.NewSyncProducer,
	InitKafka,
)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewRedisUserCache,
	repository.NewUserRepository,
	service.NewUserService)

var articlSvcProvider = wire.NewSet(
	article.NewCachedArticleRepository,
	cache.NewArticleCache,
	artdao.NewGORMArticleDAO,
	service.NewArticleService)

var interactiveSvcSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	service.NewInteractiveService,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		cache.NewRedisCodeCache,
		artdao.NewGORMArticleDAO,
		cache.NewArticleCache,
		cache.NewInteractiveRedisCache,

		article2.NewKafkaProducer,

		repository.NewCodeRepository,
		repository.NewCachedInteractiveRepository,
		article.NewCachedArticleRepository,
		ioc.InitSMSService,

		service.NewCodeService,
		service.NewArticleService,
		service.NewInteractiveService,
		dao.NewGORMInteractiveDAO,
		InitWechatService,

		web.NewArticleHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,

		InitWechatHandlerConfig,
		ijwt.NewRedisJWT,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
		//InitInteractiveService,
	)
	return gin.Default()
}

func InitArticleHandler(dao artdao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		interactiveSvcSet,
		cache.NewArticleCache,
		article.NewCachedArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,
		article2.NewKafkaProducer)
	return &web.ArticleHandler{}
}

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdPartySet, interactiveSvcSet)
	return service.NewInteractiveService(nil)
}
