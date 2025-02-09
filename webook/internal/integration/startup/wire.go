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
	artdao.NewGORMArticleDAO,
	service.NewArticleService)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articlSvcProvider,

		cache.NewRedisCodeCache,
		cache.NewArticleCache,

		article2.NewKafkaProducer,

		repository.NewCodeRepository,
		ioc.InitSMSService,

		service.NewCodeService,
		InitWechatService,

		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		web.NewUserHandler,

		InitWechatHandlerConfig,
		ijwt.NewRedisJWT,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
