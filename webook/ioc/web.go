package ioc

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"strings"
	"time"
	"webook/webook/internal/web"
	ijwt "webook/webook/internal/web/jwt"
	"webook/webook/internal/web/middleware"
	"webook/webook/pkg/ginx/middlewares/logger"
	"webook/webook/pkg/ginx/middlewares/ratelimit"
	myLogger "webook/webook/pkg/logger"
)

func InitWebServer(middlewares []gin.HandlerFunc, userHandler *web.UserHandler,
	oauth2WechatHdl *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(middlewares...)
	userHandler.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable, jwtHdl ijwt.Handler, l myLogger.LoggerV1) []gin.HandlerFunc {
	bd := logger.NewBuilder(func(ctx context.Context, al *logger.AccessLog) {
		l.Debug("http request", myLogger.Field{Key: "al", Value: al})
	}).AllowRespBody().AllowReqBody(true)
	viper.OnConfigChange(func(in fsnotify.Event) {
		ok := viper.GetBool("web.logReq")
		bd.AllowReqBody(ok)
	})
	return []gin.HandlerFunc{
		corsHandler(),
		bd.Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/users/login").
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/oauth2/wechat/authurl").
			IgnorePaths("/oauth2/wechat/callback").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/refresh_token").
			Build(),
		ratelimit.NewBuilder(NewRateLimiter(time.Second, 100)).Build(),
	}
}

func corsHandler() gin.HandlerFunc {
	return cors.New(cors.Config{
		//AllowAllOrigins: true,
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,

		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 这个是允许前端访问你的后端响应中带的头部
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
		//AllowHeaders: []string{"content-type"},
		//AllowMethods: []string{"POST"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				//if strings.Contains(origin, "localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	})
}
