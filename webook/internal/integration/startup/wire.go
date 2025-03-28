//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	article2 "github.com/TengFeiyang01/webook/webook/article/events"
	repository2 "github.com/TengFeiyang01/webook/webook/article/repository"
	cache2 "github.com/TengFeiyang01/webook/webook/article/repository/cache"
	artdao "github.com/TengFeiyang01/webook/webook/article/repository/dao"
	service2 "github.com/TengFeiyang01/webook/webook/article/service"
	"github.com/TengFeiyang01/webook/webook/internal/repository"
	"github.com/TengFeiyang01/webook/webook/internal/repository/cache"
	"github.com/TengFeiyang01/webook/webook/internal/repository/dao"
	"github.com/TengFeiyang01/webook/webook/internal/service"
	"github.com/TengFeiyang01/webook/webook/internal/web"
	ijwt "github.com/TengFeiyang01/webook/webook/internal/web/jwt"
	"github.com/TengFeiyang01/webook/webook/ioc"
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
	repository2.NewCachedArticleRepository,
	artdao.NewGORMArticleDAO,
	service2.NewArticleService)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articlSvcProvider,

		cache.NewRedisCodeCache,
		cache2.NewArticleCache,

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
