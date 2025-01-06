//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
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

func InitWebUser() *gin.Engine {
	wire.Build(
		// 初始化 DB
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,

		// 初始化 DAO
		dao.NewUserDAO,

		// 初始化 cache
		user.NewRedisUserCache,
		//cache.NewLocalCodeCache,
		cache.NewRedisCodeCache,

		// 初始化 repository
		repository.NewUserRepository,
		repository.NewCodeRepository,
		artrepo.NewCachedArticleRepository,

		// 初始化 service
		service.NewUserService,
		service.NewCodeService,

		ioc.InitSMSService,
		ioc.InitOAuth2WechatService,
		ioc.NewWechatHandlerConfig,

		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		service.NewArticleService,
		article.NewGORMArticleDAO,
		ijwt.NewRedisJWT,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
