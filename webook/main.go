package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
	"webook/webook/config"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/service/sms/memory"
	"webook/webook/internal/web"
	"webook/webook/internal/web/middleware"
)

func main() {
	db := initDB()
	rdb := initRedis()
	server := initWebServer()
	u := initUser(db, rdb)

	u.RegisterRoute(server)

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(200, "hello world")
	})
	_ = server.Run(":8080")
}

func initRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}
	if err := dao.InitTable(db); err != nil {
		panic(err)
	}
	return db
}

func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	uc := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(ud, uc)
	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(codeRepo, smsSvc)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc, codeSvc)
	return u
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	//redisClient := redis.NewClient(&redis.Options{
	//	Addr: config.Config.Redis.Addr,
	//})
	//server.Use(ratelimit.NewBuilder(redisClient, time.Minute, 100).Build())

	server.Use(cors.New(cors.Config{
		//AllowOrigins: []string{"http://localhost:3000"},
		//AllowMethods:  []string{"GET", "POST", "PUT", "PATCH"},
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		// ExposeHeaders 中的才是前端才能拿到的
		ExposeHeaders:    []string{"Content-Length", "x-jwt-token"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") || strings.Contains(origin, "dev.webook.com") {
				return true
			}
			return false
		},
		MaxAge: 12 * time.Hour,
	}))

	// 基于Cookie
	//store := cookie.NewStore([]byte("secret"))
	// 基于memstore
	//store := memstore.NewStore([]byte("llCKQpJfsx6SEEiGdeWbQTC5YgIb6vZbbNNVkJ3Im3gSpkxSxwtNOnxL1lq6WCgr"),
	//	[]byte("iGHc43BcUmP96dQQwqs1lkW6aqmGupT470Jsj4Sy5BQeyvoZjJghLluVSSwjJxxU"))
	// 基于Redis
	//store, err := redis.NewStore(16, "tcp", "localhost:6379", "", []byte("YnfSjT0y1pCwhdkMBCyLCve19jZ7xqXV"),
	//	[]byte("GpJCNEnLiNblrZj5xdY9aG5cgVdKHCxh"))
	//if err != nil {
	//	panic(err)
	//}
	//
	//server.Use(sessions.Sessions("ssid", store))
	//server.Use(middleware.NewLoginMiddlewareBuilder().
	//	IgnorePaths("/users/login").
	//	IgnorePaths("/users/signup").
	//	Build())
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/login").
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login_sms").
		IgnorePaths("/users/login_sms/code/send").
		Build())

	return server
}
