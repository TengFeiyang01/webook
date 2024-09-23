// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache/code"
	"webook/webook/internal/repository/cache/user"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	"webook/webook/internal/web/jwt"
	"webook/webook/ioc"
)

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitWebUser() *gin.Engine {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJWT(cmdable)
	loggerV1 := ioc.InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, handler, loggerV1)
	db := ioc.InitDB(loggerV1)
	userDAO := dao.NewUserDAO(db)
	userCache := user.NewRedisUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository, loggerV1)
	codeCache := code.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService(cmdable)
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, cmdable, handler)
	wechatService := ioc.InitOAuth2WechatService(loggerV1)
	wechatHandlerConfig := ioc.NewWechatHandlerConfig()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, wechatHandlerConfig, handler)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler)
	return engine
}
