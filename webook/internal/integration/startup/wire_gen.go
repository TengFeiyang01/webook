// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	article3 "webook/webook/internal/events/article"
	"webook/webook/internal/repository"
	article2 "webook/webook/internal/repository/article"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/repository/dao/article"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	"webook/webook/internal/web/jwt"
	"webook/webook/ioc"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJWT(cmdable)
	loggerV1 := InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, handler, loggerV1)
	gormDB := InitDB()
	userDAO := dao.NewUserDAO(gormDB)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository, loggerV1)
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService(cmdable)
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, cmdable, handler, loggerV1)
	wechatService := InitWechatService(loggerV1)
	wechatHandlerConfig := InitWechatHandlerConfig()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, wechatHandlerConfig, handler)
	articleDAO := article.NewGORMArticleDAO(gormDB)
	articleCache := cache.NewArticleCache(cmdable)
	articleRepository := article2.NewCachedArticleRepository(articleDAO, loggerV1, userDAO, articleCache)
	client := InitKafka()
	syncProducer := ioc.NewSyncProducer(client)
	producer := article3.NewKafkaProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer, loggerV1)
	interactiveDAO := dao.NewGORMInteractiveDAO(gormDB)
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, loggerV1, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	articleHandler := web.NewArticleHandler(articleService, loggerV1, interactiveService)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	return engine
}

func InitArticleHandler(dao2 article.ArticleDAO) *web.ArticleHandler {
	loggerV1 := InitLogger()
	gormDB := InitDB()
	userDAO := dao.NewUserDAO(gormDB)
	cmdable := ioc.InitRedis()
	articleCache := cache.NewArticleCache(cmdable)
	articleRepository := article2.NewCachedArticleRepository(dao2, loggerV1, userDAO, articleCache)
	client := InitKafka()
	syncProducer := ioc.NewSyncProducer(client)
	producer := article3.NewKafkaProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer, loggerV1)
	interactiveDAO := dao.NewGORMInteractiveDAO(gormDB)
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, loggerV1, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	articleHandler := web.NewArticleHandler(articleService, loggerV1, interactiveService)
	return articleHandler
}

func InitInteractiveService() service.InteractiveService {
	gormDB := InitDB()
	interactiveDAO := dao.NewGORMInteractiveDAO(gormDB)
	loggerV1 := InitLogger()
	cmdable := ioc.InitRedis()
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, loggerV1, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	return interactiveService
}

// wire.go:

var thirdPartySet = wire.NewSet(ioc.InitRedis, InitDB, InitLogger, ioc.NewSyncProducer, InitKafka)

var userSvcProvider = wire.NewSet(dao.NewUserDAO, cache.NewRedisUserCache, repository.NewUserRepository, service.NewUserService)

var articlSvcProvider = wire.NewSet(article2.NewCachedArticleRepository, cache.NewArticleCache, article.NewGORMArticleDAO, service.NewArticleService)

var interactiveSvcSet = wire.NewSet(dao.NewGORMInteractiveDAO, service.NewInteractiveService, cache.NewInteractiveRedisCache, repository.NewCachedInteractiveRepository)
