//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	"webook/webook/ioc"
)

func InitWebUser() *gin.Engine {
	wire.Build(
		// 初始化 DB
		ioc.InitDB, ioc.InitRedis,

		// 初始化 DAO
		dao.NewUserDAO,

		// 初始化 cache
		cache.NewUserCache,
		//cache.NewCodeMemoryCache,
		cache.NewCodeRedisCache,

		// 初始化 repository
		repository.NewUserRepository,
		repository.NewCodeRepository,

		// 初始化 service
		service.NewUserService,
		service.NewCodeService,

		ioc.InitSMSService,
		web.NewUserHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
