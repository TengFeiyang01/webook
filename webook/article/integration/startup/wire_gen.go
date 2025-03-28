// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/google/wire"
	"github.com/TengFeiyang01/webook/webook/api/proto/gen/intr/v1"
	"github.com/TengFeiyang01/webook/webook/article/events"
	"github.com/TengFeiyang01/webook/webook/article/repository"
	"github.com/TengFeiyang01/webook/webook/article/repository/cache"
	"github.com/TengFeiyang01/webook/webook/article/repository/dao"
	"github.com/TengFeiyang01/webook/webook/article/service"
	repository2 "github.com/TengFeiyang01/webook/webook/internal/repository"
	cache2 "github.com/TengFeiyang01/webook/webook/internal/repository/cache"
	dao2 "github.com/TengFeiyang01/webook/webook/internal/repository/dao"
	service2 "github.com/TengFeiyang01/webook/webook/internal/service"
	"github.com/TengFeiyang01/webook/webook/ioc"
)

// Injectors from wire.go:

func InitArticleHandler() service.ArticleService {
	gormDB := InitDB()
	articleDAO := dao.NewGORMArticleDAO(gormDB)
	loggerV1 := InitLogger()
	userDAO := dao2.NewUserDAO(gormDB)
	cmdable := InitRedis()
	articleCache := cache.NewArticleCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(articleDAO, loggerV1, userDAO, articleCache)
	client := InitKafka()
	syncProducer := ioc.NewSyncProducer(client)
	producer := events.NewKafkaProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer, loggerV1)
	return articleService
}

// wire.go:

var thirdPartySet = wire.NewSet(
	InitRedis, InitDB, InitLogger, InitKafka, ioc.NewSyncProducer,
)

var userSvcProviderSet = wire.NewSet(dao2.NewUserDAO, repository2.NewUserRepository, service2.NewUserService, cache2.NewRedisUserCache)

var articlSvcProvider = wire.NewSet(repository.NewCachedArticleRepository, dao.NewGORMArticleDAO, service.NewArticleService, events.NewKafkaProducer, intrv1.NewInteractiveServiceClient, cache.NewArticleCache)
