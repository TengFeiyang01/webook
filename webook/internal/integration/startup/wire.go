//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/article"
	"webook/webook/internal/repository/cache/code"
	"webook/webook/internal/repository/cache/user"
	"webook/webook/internal/repository/dao"
	artdao "webook/webook/internal/repository/dao/article"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	ijwt "webook/webook/internal/web/jwt"
	"webook/webook/ioc"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	ioc.InitRedis, ioc.InitDB, InitLogger)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	user.NewRedisUserCache,
	repository.NewUserRepository,
	service.NewUserService)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		// cache 部分
		code.NewRedisCodeCache,
		artdao.NewGORMArticleDAO,

		// repository 部分
		repository.NewCodeRepository,
		article.NewArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		service.NewCodeService,
		service.NewArticleService,
		InitWechatService,

		// handler 部分
		web.NewArticleHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		InitWechatHandlerConfig,
		ijwt.NewRedisJWT,

		ioc.InitGinMiddlewares,

		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler(d artdao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		//userSvcProvider,
		service.NewArticleService,
		web.NewArticleHandler,
		article.NewArticleRepository,
		//artdao.NewMongoDBArticleDAO,
	)
	return &web.ArticleHandler{}
}
