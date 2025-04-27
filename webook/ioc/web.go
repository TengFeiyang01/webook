package ioc

import (
	"github.com/TengFeiyang01/webook/webook/internal/web"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"strings"
	"time"

	ijwt "github.com/TengFeiyang01/webook/webook/internal/web/jwt"
	"github.com/TengFeiyang01/webook/webook/internal/web/middleware"
	"github.com/TengFeiyang01/webook/webook/pkg/ginx"
	"github.com/TengFeiyang01/webook/webook/pkg/ginx/middlewares/ratelimit"
	myLogger "github.com/TengFeiyang01/webook/webook/pkg/logger"
	"github.com/TengFeiyang01/webook/webook/pkg/metric"
)

func InitWebServer(middlewares []gin.HandlerFunc, userHandler *web.UserHandler,
	articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(middlewares...)
	userHandler.RegisterRoutes(server)
	//oauth2WechatHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	(&web.ObservabilityHandler{}).RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable, jwtHdl ijwt.Handler, l myLogger.LoggerV1) []gin.HandlerFunc {
	//bd := logger.NewBuilder(func(ctx context.Context, al *logger.AccessLog) {
	//	l.Debug("http request", myLogger.Field{Key: "al", Value: al})
	//}).AllowRespBody().AllowReqBody(true)
	//viper.OnConfigChange(func(in fsnotify.Event) {
	//	ok := viper.GetBool("web.logReq")
	//	bd.AllowReqBody(ok)
	//})
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "ytf",
		Subsystem: "webook",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	return []gin.HandlerFunc{
		corsHandler(),
		(&metric.MiddlewareBuilder{
			Namespace:  "ytf",
			Subsystem:  "webook",
			Name:       "gin_http",
			Help:       "统计 GIN 的 HTTP 接口",
			InstanceID: "my-instance-1",
		}).Build(),
		//bd.Build(),
		otelgin.Middleware("webook"),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/users/login").
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/oauth2/wechat/authurl").
			IgnorePaths("/oauth2/wechat/callback").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/refresh_token").
			IgnorePaths("/test/metric").
			Build(),
		ratelimit.NewBuilder(NewRateLimiter(time.Second, 100)).Build(),
	}
}

func corsHandler() gin.HandlerFunc {
	return cors.New(cors.Config{
		//AllowAllOrigins: true,
		AllowOrigins:     []string{"http://localhost:3000"},
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
