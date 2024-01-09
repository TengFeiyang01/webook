package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	"webook/webook/internal/web/middleware"
)

func main() {
	db := initDB()
	server := initWebServer()

	u := initUser(db)
	u.RegisterRoutes(server)

	server.Run(":8090")
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	//store := cookie.NewStore([]byte("secret"))

	// 这是基于内存的实现，第一个参数是 authentication key，最好是 32 或者 64 位
	// 第二个参数是 encryption key
	//store := memstore.NewStore([]byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm"),
	//	[]byte("o6jdlo2cb9f9pb6h46fjmllw481ldebj"))

	store, err := redis.NewStore(16,
		"tcp", "localhost:6379", "",
		[]byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm"), []byte("o6jdlo2cb9f9pb6h46fjmllw481ldebj"))

	if err != nil {
		panic(err)
	}

	server.Use(sessions.Sessions("mysession", store))

	server.Use(cors.New(cors.Config{
		AllowAllOrigins: false,
		AllowOrigins:    []string{"http://localhost:3000"},
		// 在使用 JWT 的时候，因为我们使用了 Authorizaition 的头部，所以需要加上
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 为了 JWT
		ExposeHeaders:    []string{"x-jwt-token", "Authorization"},
		AllowMethods:     []string{"POST", "GET", "PUT"},
		AllowCredentials: true,
		// 你不加这个 前端是拿不到的
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "abc")
		},
		MaxAge: 12 * time.Hour,
	}))

	//server.Use(middleware.NewLoginMiddleBuilder().
	//	IgnorePaths("/users/signup").
	//	IgnorePaths("/users/login").Build())

	server.Use(middleware.NewLoginJWTMiddleBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").Build())
	return server
}

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		// 我只会在初始化过程 panic
		// 一旦初始化过程出错，应用就不要启动了
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
